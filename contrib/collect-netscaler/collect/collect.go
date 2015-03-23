// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"expvar"
	"log"
	"runtime"
	"sync"
	"strings"
	"time"

	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

var (
	statCycleMillis = expvar.NewInt("collect.CycleMillis")
)

// collector represents a collector for a given subsystem.
type collector interface {
	Subsystem() string // "lbvserver", "ssl", etc.
}

// statsCollector is a collector that calls a member of the "stat" family of
// functions in the Nitro API.
type statsCollector interface {
	CollectStats(emitFn, *nitro.ResponseStat)
}

// configCollector is a collector that calls a member of the "config" family of
// functions in the Nitro API.
type configCollector interface {
	CollectConfig(emitFn, *nitro.ResponseConfig)
}

// emitFn queues a data point for emission.
type emitFn func(string, interface{})

var Client *nitro.Client

// collect emits data points based on the provided collector.
func collect(emit emitFn, c collector) {
	switch cc := c.(type) {
	default:
		log.Panicf("unsupported collector type: %T", c)

	case statsCollector:
		resp, err := Client.Stat.Get(c.Subsystem())
		if err != nil {
			log.Print(err)
			return
		}
		logPanics(func() {
			cc.CollectStats(emit, resp)
		})

	case configCollector:
		resp, err := Client.Config.Get(c.Subsystem())
		if err != nil {
			log.Print(err)
			return
		}
		logPanics(func() {
			cc.CollectConfig(emit, resp)
		})
	}
}

// Loop loops indefinitely, running all collectors at the provided interval,
// and writing to w the resulting data points.
func Loop(w chan *tsdb.Point, interval time.Duration) {
	tick := tsdb.Tick(interval)
	t := time.Now()
	for ; ; t = <-tick {
		start := time.Now()

		emit := newEmitter(w, t)
		var wg sync.WaitGroup
		for _, c := range collectors {
			c := c
			go func() {
				collect(emit, c)
				wg.Done()
			}()
			wg.Add(1)
		}
		wg.Wait()

		statCycleMillis.Add(time.Since(start).Nanoseconds() / 1e6)
	}
}

// newEmitter returns a function that emits data points for the provided
// time instant.
func newEmitter(w chan *tsdb.Point, timestamp time.Time) emitFn {
	return func(series string, value interface{}) {
		if value == nil {
			panic("zero value")
		}
		series = "netscaler." + series
		id := strings.Fields(strings.Replace(series, "=", " ", -1))
		p, err := tsdb.NewPoint(timestamp, value, id[0], id[1:]...)
		if err != nil {
			panic(err)
		}
		w <- p
	}
}

// collectors is a list of all registered collectors.
var collectors []collector

func register(c collector) {
	collectors = append(collectors, c)
}

func registerStatFunc(subsystem string, fn func(emitFn, *nitro.ResponseStat)) {
	register(statFunc{
		subsystem: subsystem,
		fn:        fn,
	})
}

func registerConfigFunc(subsystem string, fn func(emitFn, *nitro.ResponseConfig)) {
	register(configFunc{
		subsystem: subsystem,
		fn:        fn,
	})
}

// statFunc is an adapter to allow the use of ordinary functions as stats
// collectors.
type statFunc struct {
	subsystem string
	fn        func(emitFn, *nitro.ResponseStat)
}

func (sf statFunc) Subsystem() string {
	return sf.subsystem
}

func (sf statFunc) CollectStats(emit emitFn, r *nitro.ResponseStat) {
	sf.fn(emit, r)
}

// configFunc is an adapter to allow the use of ordinary functions as config
// collectors.
type configFunc struct {
	subsystem string
	fn        func(emitFn, *nitro.ResponseConfig)
}

func (cf configFunc) Subsystem() string {
	return cf.subsystem
}

func (cf configFunc) CollectConfig(emit emitFn, r *nitro.ResponseConfig) {
	cf.fn(emit, r)
}

func logPanics(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("handler panic: %v\n%s", err, buf)
		}
	}()
	fn()
}
