// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"expvar"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"opentsp.org/internal/tsdb"
)

// MaxQueue limits the number of points queued in a server.
const MaxQueue = 100000

var (
	statErrors          = expvar.NewMap("server.Errors")
	statQueue           = expvar.NewMap("server.Queue")
	statServerCurrEstab = expvar.NewInt("server.CurrEstab")
)

func ListenAndServe(addr string) <-chan *tsdb.Point {
	ch := make(chan *tsdb.Point, MaxQueue)
	s := listen(addr)
	statQueue.Set("", expvar.Func(func() interface{} {
		return len(ch)
	}))
	go s.loop(ch)
	return ch
}

func listen(addr string) *server {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return &server{l}
}

// server implements fan-in from forwarders and pollers.
type server struct {
	listener net.Listener
}

func (s *server) loop(w chan<- *tsdb.Point) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Print(err)
			time.Sleep(5 * time.Second)
			continue
		}
		c := &serverConn{Conn: conn}
		go c.loop(w)
	}
}

type serverConn struct {
	net.Conn
}

func (c *serverConn) loop(w chan<- *tsdb.Point) {
	statServerCurrEstab.Add(1)
	defer statServerCurrEstab.Add(-1)
	defer c.Close()
	var (
		r   = ProtocolReader(c)
		dec = tsdb.NewDecoder(r)
	)
	dec.DisableOrderCheck()
	for {
		point, err := dec.Decode()
		if err != nil {
			if _, ok := err.(*tsdb.SyntaxError); ok {
				statErrors.Add("type=Syntax", 1)
				fmt.Fprintf(c, "error: invalid syntax\n")
				continue
			}
			return
		}
		select {
		case w <- point:
			// ok
		default:
			statErrors.Add("type=Enqueue", 1)
			w <- point
		}
	}
}

// ProtocolReader translates puts received via the "telnet" protocol
// into ordinary io.Reader suitable for use with tsdb.Decoder.
func ProtocolReader(conn net.Conn) io.Reader {
	return &protocolReader{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}
}

type protocolReader struct {
	conn    net.Conn
	scanner *bufio.Scanner
	buf     []byte
}

var (
	putCmd     = []byte("put ")
	versionCmd = []byte("version")
)

func (pr *protocolReader) Read(buf []byte) (int, error) {
	if len(pr.buf) > 0 {
		n := copy(buf, pr.buf)
		pr.buf = pr.buf[n:]
		return n, nil
	}
	for pr.scanner.Scan() {
		cmd := pr.scanner.Bytes()
		switch {
		default:
			statErrors.Add("type=InvalidCommand", 1)
			fmt.Fprintf(pr.conn, "error: invalid command\n")
		case bytes.Equal(cmd, versionCmd):
			fmt.Fprintf(pr.conn, "Built on ... (tsp-aggregator)\n")
		case bytes.HasPrefix(cmd, putCmd):
			line := bytes.TrimPrefix(cmd, putCmd)
			line = append(line, '\n')
			n := copy(buf, line)
			if n < len(line) {
				pr.buf = append(pr.buf, line[n:]...)
			}
			return n, nil
		}
	}
	err := pr.scanner.Err()
	if err == nil {
		err = io.EOF
	}
	return 0, err
}
