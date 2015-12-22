// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package config handles the config file.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"opentsp.org/cmd/collect-statse/aggregator"
	"opentsp.org/cmd/collect-statse/forwarder"
	"opentsp.org/cmd/collect-statse/statse"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/filter"
	"opentsp.org/internal/version"

	xconfig "opentsp.org/internal/config"
)

const maxRules = 64

var Loaded *Config

func Load(tsdbChan chan<- *tsdb.Point) {
	var (
		filePath    = flag.String("f", "/etc/collect-statse/config", "configuration file")
		debugMode   = flag.Bool("v", false, "verbose mode")
		testMode    = flag.Bool("t", false, "configuration test")
		versionMode = flag.Bool("version", false, "output version and exit")
	)
	flag.Parse()
	log.SetFlags(0)
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *versionMode {
		fmt.Println(version.ToString())
		os.Exit(0)
	}
	if *debugMode {
		w := os.Stderr
		xconfig.Debug = log.New(w, "debug: config: ", 0)
		statse.Debug = log.New(w, "debug: statse: ", 0)
		forwarder.Debug = log.New(w, "debug: forwarder: ", 0)
		filter.Debug = log.New(w, "debug: forwarder/filter: ", 0)
		aggregator.Debug = log.New(w, "debug: aggregator: ", 0)
	}
	config := new(Config)
	xconfig.Load(config, *filePath, "collect-statse?host={{.Hostname}}")
	if *testMode {
		config.Dump(os.Stdout)
		os.Exit(0)
	}
	Loaded = config
	aggregator.Init(tsdbChan)
	log.Printf("start pid=%d", os.Getpid())
}

type Config struct {
	Forwarder  forwarder.Config  `config:"dynamic"`
	Aggregator aggregator.Config `config:"dynamic"`
}

func (c *Config) Reset() {
	*c = Config{}
}

func (c *Config) String() string {
	buf, _ := json.Marshal(c)
	return string(buf)
}

func (c *Config) Dump(w io.Writer) {
	buf, _ := json.MarshalIndent(c, "", "\t")
	fmt.Fprintln(w, string(buf))
}

func (c *Config) Validate() error {
	if err := c.validateForwarder(); err != nil {
		return err
	}
	if err := c.validateAggregator(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateForwarder() error {
	config := &c.Forwarder
	if err := config.Validate(); err != nil {
		return err
	}
	if n := len(config.Filter); n > maxRules {
		return fmt.Errorf("too many filter rules defined (%d>%d)", n, maxRules)
	}
	return nil
}

func (c *Config) validateAggregator() error {
	return c.Aggregator.Validate()
}

var reload = make(chan string)

func Reload() {
	go changeMonitor()
	cause := <-reload
	log.Printf("restarting... (%s)", cause)
	os.Exit(0)
}

func changeMonitor() {
	dummy := new(Config)
	xconfig.Next(dummy)
	reload <- "config updated"
}
