// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package submit

import (
	"expvar"
	"log"
	"os"
	"regexp"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/filter"
)

// serverMaxQueue limits the number of points that can be queued while
// the server is blocked. This queueing is typically due to filter complexity.
const serverMaxQueue = 10000

var statQueue = expvar.NewMap("collect.Queue")

type Server struct {
	submit chan *tsdb.Point
	filter *filter.Filter
	relay  []*Relay
}

type ServerConfig struct {
	Filter []filter.Rule
	Relay  map[string]*RelayConfig
}

func NewServer(config *ServerConfig) (*Server, error) {
	filter, err := filter.New(config.Filter...)
	if err != nil {
		return nil, err
	}
	var relay []*Relay
	for name, config := range config.Relay {
		r, err := NewRelay(name, config)
		if err != nil {
			return nil, err
		}
		relay = append(relay, r)
	}
	server := &Server{
		submit: make(chan *tsdb.Point, serverMaxQueue),
		filter: filter,
		relay:  relay,
	}
	queue := server.submit
	statQueue.Set("", expvar.Func(func() interface{} {
		return len(queue)
	}))
	return server, nil
}

func (s *Server) Submit(point *tsdb.Point) {
	s.submit <- point
}

func (s *Server) Serve() {
	var point *tsdb.Point
	var relay *Relay
	var pass bool
	var err error
	for point = range s.submit {
		pass, err = s.filter.Eval(point)
		if err != nil {
			log.Printf("filter error: %v", err)
			point.Free()
			continue
		}
		if !pass {
			point.Free()
			continue
		}
		for _, relay = range s.relay {
			relay.Submit(point)
		}
		point.Free()
	}
}

var hostRE = regexp.MustCompile("[ \t]?host=")

func hostname() string {
	s, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return s
}
