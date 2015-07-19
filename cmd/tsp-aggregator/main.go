// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// tsp-aggregator combines many host feeds into a single site feed.
package main

import (
	"flag"
	"log"
	"os"

	_ "opentsp.org/internal/pprof"

	"opentsp.org/internal/logfile"
	"opentsp.org/internal/relay"
	"opentsp.org/internal/stats"
	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/filter"
)

var (
	filePath  = flag.String("f", defaultConfigPath, "configuration file")
	debugMode = flag.Bool("v", false, "verbose mode")
	testMode  = flag.Bool("t", false, "configuration test")
)

var cfg *Config

func init() {
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	cfg = load(*filePath)
	w := logfile.Open(cfg.LogPath)
	log.SetOutput(w)
	if *debugMode {
		filter.Debug = log.New(w, "debug: filter: ", 0)
	}
	log.Print("start pid=", os.Getpid())
	go Restart()
}

func main() {
	var (
		remote = ListenAndServe(cfg.ListenAddr)
		self   = SelfStats("tsp.aggregator.", cfg.Filter)
		final  = tsdb.Join(remote, self)
		relays = relay.NewPool(cfg.Relay, final)
	)
	relays.Broadcast()
}

// SelfStats is like stats.Self except the returned tsdb.Chan is filtered using
// the given rules.
func SelfStats(prefix string, rules []filter.Rule) tsdb.Chan {
	var (
		self     = stats.Self(prefix)
		filtered = filter.Series(rules, self)
		out      = make(chan *tsdb.Point)
	)
	go func() {
		for {
			out <- filtered.Next()
		}
	}()
	return out
}
