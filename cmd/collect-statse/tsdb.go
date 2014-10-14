// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"time"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/tsdbutil"
)

func init() {
	go encode()
}

var tsdbChan = make(chan *tsdb.Point, 100000)

func encode() {
	enc := tsdbutil.NewStdoutEncoder()
	for p := range tsdbChan {
		enc.Encode(p)
	}
}

const varInterval = 10 * time.Second // expvar dump interval

func expvarLoop() {
	tick := tsdb.Tick(varInterval)
	for {
		now := <-tick
		metric := make([]byte, 0, 1024)
		tsdbutil.ExportVars(now, func(p *tsdb.Point) {
			metric = append(metric[:0], "tsp.collect-statse."...)
			metric = append(metric, p.Metric()...)
			err := p.SetMetric(metric)
			if err != nil {
				log.Panic(err)
			}
			tsdbChan <- p
		})
	}
}
