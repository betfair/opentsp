// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package control handles config requests from tsp-forwarder.
package control

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"opentsp.org/cmd/tsp-controller/config"
	"opentsp.org/cmd/tsp-controller/config/network"
	"opentsp.org/cmd/tsp-controller/control"
)

func init() {
	control.Register(Update, Handle)
}

// Update decodes the tsp-forwarder(8) parts of the config.
func Update(config *config.Config) error {
	return nil
}

// Handle registers the HTTP handlers.
func Handle(config *config.Config) error {
	handler := &handler{config}
	http.Handle("/control/v1/tsp-forwarder", handler)
	http.Handle("/control/v1/tsp-aggregator", handler)
	http.Handle("/control/v1/tsp-poller", handler)
	http.Handle("/control/v1/tsdb-forwarder", handler)  // legacy
	http.Handle("/control/v1/tsdb-aggregator", handler) // legacy
	http.Handle("/control/v1/tsdb-poller", handler)     // legacy
	return nil
}

// Key represents the configuration lookup request.
type Key struct {
	Program, Host string
}

func parseKey(url *url.URL) (*Key, error) {
	query := url.Query()
	host := query.Get("host")
	if host == "" {
		return nil, fmt.Errorf("missing query parameter: host")
	}
	query.Del("host")
	for key := range query {
		return nil, fmt.Errorf("unexpected query parameter: %s", key)
	}
	program := path.Base(url.Path)
	if strings.HasPrefix(program, "tsdb-") {
		program = strings.Replace(program, "tsdb-", "tsp-", 1)
	}
	return &Key{program, host}, nil
}

// Relay corresponds to elements of the Relay setting, see tsp-forwarder(8).
type Relay struct {
	Host            string
	DropRepeats     bool `json:",omitempty"`
	MaxConnsPerHost int  `json:",omitempty"`
}

// Rule corresponds to elements of the Filter setting, see tsp-forwarder(8).
type Rule struct {
	Match []string
	Set   []string
}

// View corresponds to tsp-forwarder configuration file, see tsp-forwarder(8).
type View struct {
	Filter []*Rule
	Relay  map[string]*Relay
}

type internalError struct{ error }

// newView returns an initial view containing common settings.
func newView(key *Key, config *config.Config) (*View, error) {
	host, err := config.Host(key.Host)
	if err != nil {
		return nil, err
	}
	view := &View{
		Relay: make(map[string]*Relay),
	}
	// Load custom filter rules.
	err = config.Filter.Run(&view.Filter, key.Program, host.ID, host.ClusterID)
	if err != nil {
		return nil, &internalError{err}
	}
	// Add host tag if missing.
	view.Filter = append(view.Filter, &Rule{
		Match: []string{"", "host", "^$"},
		Set:   []string{"", "host", host.ID},
	})
	return view, nil
}

type handler struct {
	config *config.Config
}

// View is a convenience function that calls the appropriate *View function.
func (h *handler) View(key *Key) (*View, error) {
	switch key.Program {
	default:
		panic("internal error")
	case "tsp-forwarder":
		return h.forwarderView(key)
	case "tsp-poller":
		return h.pollerView(key)
	case "tsp-aggregator":
		return h.aggregatorView(key)
	}
}

// forwarderView returns a tsp-forwarder(8) view.
func (h *handler) forwarderView(key *Key) (*View, error) {
	host, err := h.config.Host(key.Host)
	if err != nil {
		return nil, err
	}
	view, err := newView(key, h.config)
	if err != nil {
		return nil, err
	}
	// Add cluster tag if missing.
	if cluster := host.ClusterID; cluster != "" {
		view.Filter = append(view.Filter, &Rule{
			Match: []string{"", "cluster", "^$"},
			Set:   []string{"", "cluster", cluster},
		})
	}
	// Feed direct subscribers, typically just OpenTSDB.
	for _, s := range directSubscribers(h.config) {
		view.Relay[s.ID] = &Relay{
			Host:        s.Host,
			DropRepeats: s.Dedup,
		}
	}
	// Feed indirect subscribers.
	if aggregator := h.config.Network.Aggregator; aggregator != nil {
		view.Relay["aggregator"] = &Relay{
			Host: aggregator.Host,
		}
	}
	return view, nil
}

func directSubscribers(config *config.Config) []*network.Subscriber {
	return subscribers(config, func(s *network.Subscriber) bool {
		return s.Direct
	})
}

func indirectSubscribers(config *config.Config) []*network.Subscriber {
	return subscribers(config, func(s *network.Subscriber) bool {
		return !s.Direct
	})
}

func subscribers(config *config.Config, match func(*network.Subscriber) bool) []*network.Subscriber {
	var got []*network.Subscriber
	for _, s := range config.Network.Subscriber {
		if !match(s) {
			continue
		}
		got = append(got, s)
	}
	return got
}

// pollerMaxConnsPerHost is poller's limit on count of connections
// established to direct subscribers. It's motivated by OpenTSDB
// load balancing & scalability: single high-throughput connection is
// more problematic than 12 connections, each carrying 1/12 of the
// traffic.
//
// TODO(masiulaniec): make this configurable.
const pollerMaxConnsPerHost = 12

// pollerView returns a tsp-poller(8) view.
func (h *handler) pollerView(key *Key) (*View, error) {
	view, err := newView(key, h.config)
	if err != nil {
		return nil, err
	}
	// Feed direct subscribers, typically just OpenTSDB.
	for _, s := range directSubscribers(h.config) {
		view.Relay[s.ID] = &Relay{
			Host:            s.Host,
			DropRepeats:     s.Dedup,
			MaxConnsPerHost: pollerMaxConnsPerHost,
		}
	}
	// Feed indirect subscribers.
	if aggregator := h.config.Network.Aggregator; aggregator != nil {
		view.Relay["aggregator"] = &Relay{
			Host: aggregator.Host,
		}
	}
	return view, nil
}

func isAggregator(host string, config *config.Config) bool {
	aggregator := config.Network.Aggregator
	return aggregator != nil && aggregator.Host == host
}

// aggregatorView returns a tsp-aggregator(8) view.
func (h *handler) aggregatorView(key *Key) (*View, error) {
	if !isAggregator(key.Host, h.config) {
		return nil, fmt.Errorf("not an aggregator: %v", key.Host)
	}
	view, err := newView(key, h.config)
	if err != nil {
		return nil, err
	}
	// Feed indirect subscribers.
	for _, s := range indirectSubscribers(h.config) {
		view.Relay[s.ID] = &Relay{
			Host:        s.Host,
			DropRepeats: s.Dedup,
		}
	}
	return view, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key, err := parseKey(req.URL)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	view, err := h.View(key)
	if err != nil {
		log.Printf("%s: %v", key.Program, err)
		if _, ok := err.(*internalError); ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	control.Marshal(w, view)
}
