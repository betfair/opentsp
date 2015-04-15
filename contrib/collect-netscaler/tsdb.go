// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"os"
	"time"

	"opentsp.org/contrib/collect-netscaler/config"
	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/tsdbutil"
)

func init() {
	go encode()
}

var tsdbChan = make(chan *tsdb.Point)

func encode() {
	enc := tsdb.NewEncoder(os.Stdout)
	for p := range tsdbChan {
		p.XAppendTags("host", tsdb.Clean(config.Host()))
		if err := enc.Encode(p); err != nil {
			log.Fatal(err)
		}
	}
}

func expvarLoop() {
	tick := tsdb.Tick(10 * time.Second)
	for {
		t := <-tick
		tsdbutil.ExportVars(t, func(p *tsdb.Point) {
			metric := []byte("tsp.collect-netscaler.")
			metric = append(metric, p.Metric()...)
			if err := p.SetMetric(metric); err != nil {
				panic(err)
			}
			tsdbChan <- p
		})
	}
}

func init() {
	go expvarLoop()
}
