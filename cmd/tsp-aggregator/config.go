package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"opentsp.org/internal/config"
	"opentsp.org/internal/flag"
	"opentsp.org/internal/relay"
	"opentsp.org/internal/restart"
	"opentsp.org/internal/tsdb/filter"
	"opentsp.org/internal/validate"
)

type Config struct {
	Filter     []filter.Rule            `config:"dynamic"`
	ListenAddr string                   `config:"dynamic"`
	Relay      map[string]*relay.Config `config:"dynamic"`
	LogPath    string
}

func load(path string) *Config {
	if flag.DebugMode {
		config.Debug = log.New(os.Stderr, "debug: config: ", 0)
	}
	cfg := new(Config)
	config.Load(cfg, path, "tsp-aggregator?host={{.Hostname}}")
	if flag.TestMode {
		cfg.Dump(os.Stdout)
		os.Exit(0)
	}
	go func() {
		dummy := new(Config)
		config.Next(dummy)
		restartCause <- "config updated"
	}()
	return cfg
}

func (c *Config) Reset() { *c = *defaultConfig }

func (c *Config) Dump(w io.Writer) {
	buf, _ := json.MarshalIndent(c, "", "\t")
	fmt.Fprintln(w, string(buf))
}

func (c *Config) Validate() error {
	var err error
	c.Filter, err = validate.Filter(c.Filter)
	if err != nil {
		return err
	}
	if err := validate.Relay(c.Relay); err != nil {
		return err
	}
	if err := validate.ListenAddr(c.ListenAddr); err != nil {
		return err
	}
	return nil
}

var restartCause = make(chan string)

func Restart() {
	cause := <-restartCause
	log.Printf("restarting... (%s)", cause)
	restart.Do()
}
