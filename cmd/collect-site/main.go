// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"expvar"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/tsdbutil"
)

var listenAddr = flag.String("l", ":4242", "listen address")

var (
	output   io.Writer = os.Stdout
	outputMu sync.Mutex
)

var (
	statServerCurrEstab = expvar.NewInt("server.CurrEstab")
)

func init() {
	log.SetFlags(0)
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	go submitVars()
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	server := &Server{Listener: l}
	server.Serve()
}

type Server struct {
	Listener net.Listener
}

func (s *Server) Serve() {
	for {
		clientConn, err := s.Listener.Accept()
		if err != nil {
			log.Print(err)
			time.Sleep(5 * time.Second)
			continue
		}
		c := &conn{Conn: clientConn}
		go c.serve()
	}
}

type conn struct {
	net.Conn
}

func (c *conn) serve() {
	var (
		putCmd      = []byte("put ")
		versionCmd  = []byte("version")
		versionRate = time.NewTicker(1 * time.Second)
		scanner     = bufio.NewScanner(c)
		cmd, point  []byte
	)
	defer versionRate.Stop()
	statServerCurrEstab.Add(1)
	defer statServerCurrEstab.Add(-1)
	for scanner.Scan() {
		cmd = scanner.Bytes()
		switch {
		default:
			// ignore unexpected commands
		case bytes.HasPrefix(cmd, versionCmd):
			<-versionRate.C
			c.Write([]byte("Built on ... (collect-site)\n"))
		case bytes.HasPrefix(cmd, putCmd):
			point = bytes.TrimPrefix(cmd, putCmd)
			submit(point)
		}
	}
	_ = scanner.Err()
}

// varInterval is the interval between dumps of exported vars.
const varInterval = 1 * time.Second

func submitVars() {
	enc := tsdb.NewEncoder(output)
	tick := tsdb.Tick(varInterval)
	for {
		now := <-tick
		metric := make([]byte, 0, 1024)
		tsdbutil.ExportVars(now, func(point *tsdb.Point) {
			metric = append(metric[:0], "tsp.collect-site."...)
			metric = append(metric, point.Metric()...)
			err := point.SetMetric(metric)
			if err != nil {
				log.Panic(err)
			}
			outputMu.Lock()
			enc.Encode(point)
			outputMu.Unlock()
		})
	}
}

func submit(point []byte) {
	point = append(point, '\n')
	outputMu.Lock()
	if _, err := output.Write(point); err != nil {
		log.Fatal(err)
	}
	outputMu.Unlock()
}
