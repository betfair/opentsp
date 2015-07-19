// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// tsp-poller is like tsp-forwarder except it extracts remote data.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	_ "opentsp.org/internal/pprof"

	"opentsp.org/internal/collect"
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
	w := openLog(cfg.LogPath)
	log.SetOutput(w)
	if *debugMode {
		collect.Debug = log.New(w, "debug: collect: ", 0)
		filter.Debug = log.New(w, "debug: filter: ", 0)
	}
	log.Print("start pid=", os.Getpid())
}

func main() {
	var (
		plugins = collect.NewPool(cfg.CollectPath)
		self    = stats.Self("tsp.poller.")
		joined  = tsdb.Join(plugins.C, self)
		final   = filter.Series(cfg.Filter, joined)
		relays  = relay.NewPool(cfg.Relay, final)
	)
	go Restart(func() {
		plugins.Kill()
	})
	relays.Broadcast()
}

func openLog(path string) io.Writer {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
