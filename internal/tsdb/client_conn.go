// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"bufio"
	"bytes"
	"expvar"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

const (
	clientConnDefaultPort = "4242"
	clientConnDialTimeout = 15 * time.Second
	clientConnMaxSleep    = 10 * time.Minute
	clientConnAckInterval = 5 * time.Second
	clientConnAckTimeout  = 5 * time.Second
)

var statClientErrors = expvar.NewMap("tsdb.client.Errors")

// clientConn represents a single connection to TSDB server.
type clientConn struct {
	io.Closer
	Addr    string
	Pending bytes.Buffer
	r       *bufio.Scanner
	w       io.Writer
	ackTime time.Time
}

// newDial returns a dial function.
//
// Requirements: (1) first attempt must be without delay, (2) subsequent attempts
// must be delayed exponentially with some fuzz, (3) connection must be tested for
// basic sanity using the version command.
func newDial() func(string) (net.Conn, error) {
	m, mu := make(map[string]func()), sync.Mutex{}
	return func(addr string) (net.Conn, error) {
		addr = addrFull(addr)
		mu.Lock()
		retrySleep := m[addr]
		if retrySleep == nil {
			m[addr] = newRetrySleep()
			mu.Unlock()
		} else {
			mu.Unlock()
			retrySleep()
		}
		tcp, err := net.DialTimeout("tcp", addr, clientConnDialTimeout)
		if err != nil {
			return nil, err
		}
		if err := tcp.(*net.TCPConn).SetNoDelay(false); err != nil {
			tcp.Close()
			return nil, err
		}
		conn := newClientConn(tcp, addr)
		if err := conn.ack(); err != nil {
			tcp.Close()
			return nil, err
		}
		mu.Lock()
		delete(m, addr)
		mu.Unlock()
		return tcp, nil
	}
}

var (
	dialRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
	dialRandMu sync.Mutex
)

// dialExpDelay returns a number in range 1.0-2.0 for use in calculation of
// exponential delay in case of dial errors.
func dialExpDelay() float32 {
	dialRandMu.Lock()
	defer dialRandMu.Unlock()
	return 1 + dialRand.Float32()
}

func newRetrySleep() func() {
	delay := time.Duration(1 * time.Second)
	return func() {
		delay = time.Duration(float32(delay) * dialExpDelay())
		if max := clientConnMaxSleep; delay > max {
			delay = max
		}
		time.Sleep(delay)
	}
}

func newClientConn(conn net.Conn, addr string) *clientConn {
	return &clientConn{
		Addr:   addr,
		r:      bufio.NewScanner(conn),
		w:      conn,
		Closer: conn,
	}
}

func (conn *clientConn) Put(cmd cmd) error {
	if err := conn.put(cmd); err != nil {
		return err
	}
	if err := conn.ack(); err != nil {
		return err
	}
	return nil
}

// put sends a put request. There is no immediate ack; multiple puts are
// batch-acked in version.
func (conn *clientConn) put(cmd cmd) error {
	if _, err := conn.w.Write(cmd); err != nil {
		statClientErrors.Add("type=Network", 1)
		return err
	}
	// Keep a copy in case the connection breaks.
	pointBuf := cmd.Point()
	conn.Pending.Write(pointBuf)
	conn.Pending.WriteString("\n")
	// Update encoding stats; Client is effectively a network encoder.
	statEncoderBytes.Add(int64(len(pointBuf)) + 1)
	cmd.Free()
	return nil
}

func (conn *clientConn) ack() error {
	now := time.Now()
	if now.Sub(conn.ackTime) < clientConnAckInterval {
		return nil
	}
	if err := conn.version(); err != nil {
		return err
	}
	conn.ackTime = now
	conn.Pending.Reset()
	return nil
}

// version sends a version request and awaits its response.
func (conn *clientConn) version() error {
	conn.w.Write([]byte("version\n"))
	// TODO(masiulaniecj): stopping and waiting for version response leads
	// to temporarily reduced throughput. If this becomes a problem, replace
	// stop-and-wait with a windowed approach, for example: allow N version
	// requests in flight.
	if err := conn.readVersionResponse(); err != nil {
		return err
	}
	return nil
}

func (conn *clientConn) readVersionResponse() error {
	scanner := conn.r
Again:
	deadline := time.Now().Add(clientConnAckTimeout)
	// TODO(masiulaniecj): replace SetReadDeadline with Time.After.
	if err := conn.Closer.(net.Conn).SetReadDeadline(deadline); err != nil {
		return err
	}
	var resp []byte
	if scanner.Scan() {
		resp = scanner.Bytes()
		switch {
		default:
			// Server error, likely related to past put request.
			statClientErrors.Add("type=Server", 1)
		case isVersionResponsePart1(resp):
			// ok
		case isVersionResponsePart2(resp):
			return nil
		}
		goto Again
	}
	err := scanner.Err()
	if err == nil {
		err = io.EOF
	}
	statClientErrors.Add("type=Network", 1)
	return err
}

var versionResponsePart1 = []byte("net.opentsdb built at revision ")

func isVersionResponsePart1(s []byte) bool {
	return bytes.HasPrefix(s, versionResponsePart1)
}

var versionResponsePart2 = []byte("Built on ")

func isVersionResponsePart2(s []byte) bool {
	return bytes.HasPrefix(s, versionResponsePart2)
}

func addrFull(s string) string {
	_, _, err := net.SplitHostPort(s)
	if err != nil {
		return net.JoinHostPort(s, clientConnDefaultPort)
	}
	return s
}
