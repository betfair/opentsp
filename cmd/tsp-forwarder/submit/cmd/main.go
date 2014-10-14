// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package cmd

import (
	"log"
	"time"

	"opentsp.org/cmd/tsp-forwarder/collect"
	"opentsp.org/cmd/tsp-forwarder/config"
	"opentsp.org/cmd/tsp-forwarder/submit"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/tsdbutil"

	_ "opentsp.org/internal/pprof"
)

// Run runs the forwarding engine under a given program name.
func Run(program string) {
	config.Load(program)
	server, err := submit.NewServer(&submit.ServerConfig{
		Filter: config.Loaded.Filter,
		Relay:  config.Loaded.Relay,
	})
	if err != nil {
		log.Fatal(err)
	}
	pool := collect.NewPool(config.Loaded.CollectPath, server)
	go config.Reload(program, func() {
		pool.Kill()
	})
	go server.Serve()
	go collectVars(program, server)
	select {}
}

const varInterval = 10 * time.Second // expvar dump interval

func collectVars(program string, s collect.Submitter) {
	tick := tsdb.Tick(varInterval)
	for {
		now := <-tick
		metric := make([]byte, 0, 1024)
		tsdbutil.ExportVars(now, func(p *tsdb.Point) {
			metric = append(metric[:0], "tsp."...)
			metric = append(metric, program...)
			metric = append(metric, '.')
			metric = append(metric, p.Metric()...)
			err := p.SetMetric(metric)
			if err != nil {
				log.Panic(err)
			}
			s.Submit(p)
		})
	}
}
