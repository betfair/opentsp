// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package forwarder forwards events to the aggregator.
package forwarder

import (
	"expvar"
	"time"

	"opentsp.org/cmd/collect-statse/aggregator"
	"opentsp.org/cmd/collect-statse/statse"

	"opentsp.org/internal/tsdb/filter"
)

var statServerErrors = expvar.NewMap("forwarder.server.Errors")

var defaultServer *server

// ListenAndServe starts Statse listener that forwards incoming Statse messages
// to the aggregator.
func ListenAndServe(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	filter, err := filter.New(config.Filter...)
	if err != nil {
		return err
	}
	l, err := statse.Listen(statse.ListenAddr)
	if err != nil {
		return err
	}
	defaultServer = &server{
		dec:        statse.NewDecoder(l),
		filter:     filter,
		aggregator: aggregator.NewClient(config.AggregatorHost),
	}
	return defaultServer.loop()
}

type server struct {
	dec        *statse.Decoder
	filter     *filter.Filter
	aggregator *aggregator.Client
}

func (s *server) loop() error {
	var event Event
	for {
		if err := s.in(&event); err != nil {
			return err
		}
		if !s.test(&event) {
			continue
		}
		s.out(event.Final())
	}
}

func (s *server) in(event *Event) error {
	if err := s.dec.Decode(&event.Statse); err != nil {
		return err
	}
	if Debug != nil {
		Debug.Printf("receive %v, delay=%v", &event.Statse, time.Since(event.Statse.Time))
	}
	event.Reset()
	return nil
}

func (s *server) test(event *Event) bool {
	pass, err := s.filter.Eval(event)
	if err != nil {
		statServerErrors.Add("type=Filter", 1)
		if Debug != nil {
			Debug.Print("drop unfilterable ", event)
		}
		return false
	}
	if !pass {
		if Debug != nil {
			Debug.Print("drop blocked ", event)
		}
		return false
	}
	return true
}

func (s *server) out(event *statse.Event) {
	if Debug != nil {
		Debug.Printf("send %v, delay=%v", event, time.Since(event.Time))
	}
	s.aggregator.Send(event)
}
