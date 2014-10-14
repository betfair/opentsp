// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"encoding/gob"
	"expvar"
	"io"
	"log"
	"net"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

var (
	statClientErrors = expvar.NewMap("aggregator.client.Errors")
	statClientBytes  = expvar.NewInt("aggregator.client.Bytes")
)

// Client represents a client of the Aggregator service.
type Client struct {
	host string
	dial dialFunc
	conn *clientConn
}

// NewClient allocates a new client to the Aggregator service.
func NewClient(host string) *Client {
	return &Client{
		host: host,
		dial: dial,
	}
}

func (c *Client) Close() {
	c.conn.Close()
}

// Send sends an event to the Aggregator service.
func (c *Client) Send(event *statse.Event) {
	if c.conn == nil {
		c.conn = c.dial(c.host)
	}
	if err := c.conn.Encode(event); err != nil {
		if c.handle(err) {
			return
		}
		log.Panicf("aggregator.Client: internal error: %v", err)
	}
}

func (c *Client) handle(err error) bool {
	if _, ok := err.(net.Error); !ok {
		return false
	}
	log.Printf("aggregator: client: %v, host=%v", err, c.host)
	statClientErrors.Add("type=Network", 1)
	c.conn.Close()
	c.conn = nil
	return true
}

type dialFunc func(string) *clientConn

func dial(host string) *clientConn {
	var conn net.Conn
	for {
		var err error
		conn, err = net.Dial("tcp", host+":"+listenPort)
		if err != nil {
			log.Printf("aggregator: client: %v, host=%v", err, host)
			statClientErrors.Add("type=Network", 1)
			time.Sleep(5 * time.Second)
			continue
		}
		if err := conn.(*net.TCPConn).SetNoDelay(false); err != nil {
			log.Printf("aggregator: client: %v", err)
		}
		break
	}
	w := &expvarClientBytes{conn}
	return &clientConn{
		Encoder: gob.NewEncoder(w),
		Writer:  w,
		Closer:  conn,
	}
}

type expvarClientBytes struct {
	net.Conn
}

func (b *expvarClientBytes) Write(buf []byte) (int, error) {
	n, err := b.Conn.Write(buf)
	if err == nil {
		statClientBytes.Add(int64(n))
	}
	return n, err
}

type clientConn struct {
	*gob.Encoder
	io.Writer
	io.Closer
}
