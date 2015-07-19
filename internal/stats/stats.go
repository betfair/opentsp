// Package stats converts expvar counters into time series.
package stats

import (
	"log"
	"time"

	"opentsp.org/internal/tsdb"
	"opentsp.org/internal/tsdb/tsdbutil"
)

const Interval = 10 * time.Second

// Self returns a time series that carries complete dump of expvar
// variables. Refreshed according to Interval.
func Self(prefix string) <-chan *tsdb.Point {
	ch := make(chan *tsdb.Point)
	go func() {
		tick := tsdb.Tick(Interval)
		for {
			now := <-tick
			metric := make([]byte, 0, 1024)
			tsdbutil.ExportVars(now, func(p *tsdb.Point) {
				metric = append(metric[:0], prefix...)
				metric = append(metric, p.Metric()...)
				err := p.SetMetric(metric)
				if err != nil {
					log.Panic(err)
				}
				ch <- p
			})
		}
	}()
	return ch
}
