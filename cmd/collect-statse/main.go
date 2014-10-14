// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// collect-statse implements continuous statistical aggregation of cluster event metrics.
package main

import (
	"log"

	"opentsp.org/cmd/collect-statse/aggregator"
	"opentsp.org/cmd/collect-statse/config"
	"opentsp.org/cmd/collect-statse/forwarder"

	_ "opentsp.org/internal/pprof"
)

func main() {
	config.Load(tsdbChan)
	go forwarderService()
	go aggregatorService()
	go config.Reload()
	go expvarLoop()
	select {}
}

func forwarderService() {
	err := forwarder.ListenAndServe(&config.Loaded.Forwarder)
	if err != nil {
		log.Fatal(err)
	}
}

func aggregatorService() {
	err := aggregator.ListenAndServe(&config.Loaded.Aggregator)
	if err != nil {
		log.Fatal(err)
	}
}
