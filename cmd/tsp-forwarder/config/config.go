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

	"opentsp.org/cmd/tsp-forwarder/collect"
	"opentsp.org/cmd/tsp-forwarder/submit"

	"opentsp.org/internal/tsdb/filter"

	xconfig "opentsp.org/internal/config"
	xrestart "opentsp.org/internal/restart"
)

const (
	maxRules  = 64
	maxRelays = 8
)

var Loaded *Config

func Load(program string) {
	var (
		filePath  = flag.String("f", defaultByProgram[program].FilePath, "configuration file")
		debugMode = flag.Bool("v", false, "verbose mode")
		testMode  = flag.Bool("t", false, "configuration test")
	)
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *debugMode {
		xconfig.Debug = log.New(os.Stderr, "debug: config: ", 0)
	}
	config := newConfig(program)
	xconfig.Load(config, *filePath, "tsp-"+program+"?host={{.Hostname}}")
	if *testMode {
		config.Dump(os.Stdout)
		os.Exit(0)
	}
	w := newOutput(config.LogPath)
	log.SetOutput(w)
	if *debugMode {
		collect.Debug = log.New(w, "debug: collect: ", 0)
		filter.Debug = log.New(w, "debug: filter: ", 0)
	}
	log.Print("start",
		" pid=", os.Getpid(),
	)
	Loaded = config
}

func newOutput(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

type Config struct {
	FilePath    string        `json:"-"`
	Filter      []filter.Rule `config:"dynamic"`
	LogPath     string
	Relay       map[string]*submit.RelayConfig `config:"dynamic"`
	CollectPath string
	reset       *Config
}

func newConfig(program string) *Config {
	return &Config{
		reset: defaultByProgram[program],
	}
}

func (c *Config) Reset() {
	reset := c.reset
	*c = *reset
	c.reset = reset
}

func (c *Config) Dump(w io.Writer) {
	buf, _ := json.MarshalIndent(c, "", "\t")
	fmt.Fprintln(w, string(buf))
}

func (c *Config) Validate() error {
	if err := c.validateFilter(); err != nil {
		return err
	}
	if err := c.validateRelay(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateFilter() error {
	if c.Filter == nil {
		c.Filter = []filter.Rule{{Block: true}}
		return nil
	}
	if n := len(c.Filter); n > maxRules {
		return fmt.Errorf("too many filter rules defined (%d>%d)", n, maxRules)
	}
	_, err := filter.New(c.Filter...)
	if err != nil {
		return fmt.Errorf("error creating filter: %v", err)
	}
	return nil
}

func (c *Config) validateRelay() error {
	if c.Relay == nil {
		return fmt.Errorf("missing setting: Relay")
	}
	if n := len(c.Relay); n > maxRelays {
		return fmt.Errorf("too many relays defined: %d > %d", n, maxRelays)
	}
	for _, config := range c.Relay {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	return nil
}

var reload = make(chan string)

func Reload(program string, stop func()) {
	go changeMonitor(program)
	cause := <-reload
	log.Printf("restarting... (%s)", cause)
	stop()
	xrestart.Do()
}

func changeMonitor(program string) {
	dummy := newConfig(program)
	xconfig.Next(dummy)
	reload <- "config updated"
}
