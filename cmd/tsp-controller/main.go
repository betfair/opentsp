// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// tsp-controller provides simplified configuration interface.
package main

import (
	"log"

	"opentsp.org/cmd/tsp-controller/config"
	"opentsp.org/cmd/tsp-controller/control"

	_ "opentsp.org/cmd/tsp-controller/control/collect-jmx"
	_ "opentsp.org/cmd/tsp-controller/control/collect-statse"
	_ "opentsp.org/cmd/tsp-controller/control/tsp-forwarder"
)

func main() {
	config.Load()
	err := control.ListenAndServe(*config.ListenAddr, config.Loaded)
	log.Fatal(err)
}
