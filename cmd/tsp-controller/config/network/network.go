// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package network handles the network config file.
package network

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

var (
	// DefaultConfig is the default configuration, served in absence of the config file.
	DefaultConfig = defaultConfig()

	// DefaultRestrictions is the default resrict ruleset of Config. It is effective in absence
	// of ruleset in the config file, or in absence of the config file itself. In this mode,
	// controller's scope is maximal: it handles every declared host.
	DefaultRestrictions = []*Restriction{
		{Host: regexp.MustCompile(".")},
	}
)

func defaultConfig() Config {
	self, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return Config{
		restrict: DefaultRestrictions,
		Subscriber: []*Subscriber{
			{ID: "default", Host: self, Direct: true, Dedup: true},
		},
	}
}

// Config contains network configuration details.
type Config struct {
	// Restrict is a set of restrictions that reduce the scope of controller to a
	// subset of hosts declared in the main config file.
	restrict []*Restriction

	// Network topology.
	Aggregator *Aggregator
	Poller     *Poller
	Subscriber []*Subscriber
}

// InScope reports if the given hostname is in scope.
func (c *Config) InScope(host string) bool {
	if len(c.restrict) == 0 {
		panic("invalid Config")
	}
	for _, rule := range c.restrict {
		if rule.Host.MatchString(host) {
			return true
		}
	}
	return false
}

func (c *Config) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	path, err := unmarshalPath(dec, start)
	if err != nil {
		return err
	}
	config, err := ReadFile(path)
	if err != nil {
		return err
	}
	*c = *config
	return nil
}

func ReadFile(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DefaultConfig, nil
		}
		return nil, err
	}
	var f struct {
		Restrict   []*Restriction `xml:"restrict"`
		Aggregator *Aggregator    `xml:"aggregator"`
		Poller     *Poller        `xml:"poller"`
		Subscriber []*Subscriber  `xml:"subscriber"`
	}
	r := bytes.NewBuffer(buf)
	if err := xml.NewDecoder(r).Decode(&f); err != nil {
		return nil, fmt.Errorf("error decoding %s: %v", path, err)
	}
	config := Config{f.Restrict, f.Aggregator, f.Poller, f.Subscriber}
	if len(config.restrict) == 0 {
		config.restrict = DefaultRestrictions
	}
	return &config, nil
}

func unmarshalPath(dec *xml.Decoder, start xml.StartElement) (string, error) {
	var config struct {
		Path string `xml:"path,attr"`
	}
	if err := dec.DecodeElement(&config, &start); err != nil {
		return "", err
	}
	path := config.Path
	if path == "" {
		path = DefaultPath
	}
	return path, nil
}

// A Restriction is a rule used to reduce scope of the controller.
type Restriction struct {
	Host *regexp.Regexp
}

func (r *Restriction) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var elem struct {
		Host string `xml:"host,attr"`
	}
	if err := dec.DecodeElement(&elem, &start); err != nil {
		return err
	}
	re, err := regexp.Compile(elem.Host)
	if err != nil {
		return err
	}
	*r = Restriction{re}
	return nil
}

// Aggregator specifies the aggregator node, a node that exports the
// indirect site feed.
type Aggregator struct {
	Host string `xml:"host,attr"`
}

// Poller represents the poller node, a node that masquerades many remote devices.
type Poller struct {
	Host string `xml:"host,attr"`
}

// Subscriber represents a consumer of the site feed. The feed arrives in the aggregated
// form (unless Direct is true), and without any pre-processing (unless Dedup is true).
type Subscriber struct {
	ID     string `xml:"id,attr"`
	Host   string `xml:"host,attr"`
	Direct bool   `xml:"direct,attr"`
	Dedup  bool   `xml:"dedup,attr"`
}
