// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package control implements control handlers for collect-statse(8)
package control

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"opentsp.org/cmd/tsp-controller/config"
	"opentsp.org/cmd/tsp-controller/control"
)

func init() {
	control.Register(Update, Handle)
}

// Update decodes the collect-statse(8) parts of the config.
func Update(config_ *config.Config) error {
	return config_.UpdateHost(func(hostID string, elem *config.Element) error {
		hc := new(HostConfig)
		if err := xml.Unmarshal(elem.Raw, hc); err != nil {
			return nil
		}
		if err := hc.validate(); err != nil {
			return fmt.Errorf("host %s: statse: %v", hostID, err)
		}
		elem.Value = hc
		return nil
	})
}

// Handle installs the HTTP handler.
func Handle(config *config.Config) error {
	handler := &handler{config}
	http.Handle("/control/v1/collect-statse", handler)
	return nil
}

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
	return &Key{"collect-statse", host}, nil
}

type Rule struct {
	Match []string `json:",omitempty"`
	Set   []string `json:",omitempty"`
	Block bool     `json:",omitempty"`
}

type View struct {
	Forwarder  ForwarderConfig
	Aggregator *AggregatorConfig `json:",omitempty"`
}

type ForwarderConfig struct {
	AggregatorHost string  `json:",omitempty"`
	Filter         []*Rule `json:",omitempty"`
}

type AggregatorConfig struct {
	SnapshotInterval string `json:",omitempty"`
}

type HostConfig struct {
	XMLName    xml.Name `xml:"statse"`
	Aggregator bool     `xml:"aggregator,attr"`
	Interval   string   `xml:"interval,attr"`
}

func (hc *HostConfig) validate() error {
	if !hc.Aggregator && hc.Interval != "" {
		return fmt.Errorf("interval set for non-aggregator")
	}
	if hc.Interval == "" {
		hc.Interval = "10s"
	}
	return nil
}

type handler struct {
	config *config.Config
}

type internalError struct {
	error
}

// View returns a collect-statse(8) view.
func (h *handler) View(key *Key) (*View, error) {
	host, err := h.config.Host(key.Host)
	if err != nil {
		return nil, err
	}
	view := new(View)
	// Load custom filter rules.
	err = h.config.Filter.Run(&view.Forwarder.Filter, key.Program, host.ID, host.ClusterID)
	if err != nil {
		return nil, &internalError{err}
	}
	// Add host tag if missing.
	view.Forwarder.Filter = append(view.Forwarder.Filter, &Rule{
		Match: []string{"", "host", "^$"},
		Set:   []string{"", "host", key.Host},
	})
	// Add cluster tag if missing.
	if cluster := host.ClusterID; cluster != "" {
		view.Forwarder.Filter = append(view.Forwarder.Filter, &Rule{
			Match: []string{"", "cluster", "^$"},
			Set:   []string{"", "cluster", cluster},
		})
	}
	// Set aggregator host.
	if ahost, ok := aggregatorHost(host, h.config); ok {
		view.Forwarder.AggregatorHost = ahost
	}
	// Set snapshot interval.
	if interval, ok := snapshotInterval(host, h.config); ok {
		view.Aggregator = &AggregatorConfig{interval}
	}
	return view, nil
}

func aggregatorHost(host *config.Host, config *config.Config) (string, bool) {
	aggregator, ok := aggregator(host, config)
	if !ok {
		return "", false
	}
	return aggregator.ID, true
}

func snapshotInterval(host *config.Host, config *config.Config) (string, bool) {
	aggregator, ok := aggregator(host, config)
	if !ok {
		return "", false
	}
	// Don't serve the setting to forwarder-only hosts.
	if host.ID != aggregator.ID {
		return "", false
	}
	hc, _ := hostConfig(aggregator)
	return hc.Interval, true
}

// aggregator returns the host responsible for aggregation of flows
// for the given forwarder host.
func aggregator(forwarder *config.Host, config *config.Config) (*config.Host, bool) {
	for _, host := range config.Cluster(forwarder.ClusterID) {
		hc, ok := hostConfig(host)
		if !ok {
			continue
		}
		if hc.Aggregator {
			return host, true
		}
	}
	return nil, false
}

func hostConfig(host *config.Host) (*HostConfig, bool) {
	for _, elem := range host.Extra {
		if hc, ok := elem.Value.(*HostConfig); ok {
			return hc, ok
		}
	}
	return nil, false
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
