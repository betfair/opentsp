// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"encoding/gob"
	"expvar"
	"fmt"
	"net"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

const listenPort = "14445"

var (
	statServerErrors  = expvar.NewMap("aggregator.server.Errors")
	statServerRecords = expvar.NewMap("aggregator.server.Records")
)

type server struct {
	listener  net.Listener
	store     *store
	scheduler *scheduler
}

// ListenAndServe starts a listener that stores incoming messages in a bounded
// buffer, and periodically summarises them by creating data points.
func ListenAndServe(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	l, err := net.Listen("tcp", "0.0.0.0:"+listenPort)
	if err != nil {
		return err
	}
	store := newStore()
	snapshotInterval, _ := time.ParseDuration(config.SnapshotInterval)
	server := &server{
		listener:  l,
		store:     store,
		scheduler: newScheduler(store, snapshotInterval),
	}
	return server.loop()
}

func (s *server) loop() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		serverConn := &serverConn{conn, s.store}
		go serverConn.loop()
	}
}

type serverConn struct {
	net.Conn
	Store *store
}

func (conn *serverConn) loop() {
	defer conn.Close()
	var (
		dec = gob.NewDecoder(conn)
		key = fmt.Sprintf("addr=%s", conn.RemoteAddr().(*net.TCPAddr).IP)
	)
	event := new(statse.Event)
	for {
		*event = statse.Event{}
		if err := dec.Decode(event); err != nil {
			conn.die(err)
			return
		}
		if Debug != nil {
			Debug.Printf("receive %v, delay=%v", event, time.Since(event.Time))
		}
		conn.Store.Write(event)
		statServerRecords.Add(key, 1)
	}
}

func (conn *serverConn) die(err error) {
	if Debug != nil {
		Debug.Printf("serverConn: %v, client=%v", err, conn.RemoteAddr())
	}
	if _, ok := err.(net.Error); ok {
		statServerErrors.Add("type=Network", 1)
	} else {
		statServerErrors.Add("type=Decode", 1)
	}
}
