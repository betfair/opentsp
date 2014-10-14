// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"expvar"
	"log"
	"time"
)

var (
	statTickerErrors = expvar.NewMap("tsdb.ticker.Errors")
)

// A Ticker holds a channel that delivers `ticks' of a clock at intervals.
type Ticker struct {
	C      <-chan time.Time
	stop   chan bool
	ticker time.Ticker
}

// NewTicker acts like time.NewTicker except the ticks are specially aligned
// in time so that the risk of generating overlapping data points is minimised.
// The ticks' clock resolution is reduced to 1 second, the maximum supported
// by TSDB.
func NewTicker(d time.Duration) *Ticker {
	if d < maxTimePrecision {
		log.Panicln("duration too short:", d)
	}
	out := make(chan time.Time)
	t := &Ticker{
		C:    out,
		stop: make(chan bool),
	}
	go t.mainloop(d, out)
	return t
}

func (t *Ticker) mainloop(d time.Duration, out chan time.Time) {
	// Even under normal conditions, some ticks will arrive few
	// milliseconds early or late. Under CPU starvation, this effect
	// will be exacerbated. To minimise risk of lost poll cycles,
	// try to align ticker to the middle of the second to give deviations
	// in each direction equal room (500ms each way). This also helps
	// getting consistent behaviour across restarts.
	now := time.Now()
	start := now.Truncate(1 * time.Second).Add(500 * time.Millisecond)
	if start.Before(now) {
		start.Add(1 * time.Second)
	}
	time.Sleep(start.Sub(now))
	tick := time.NewTicker(d)
	defer tick.Stop()

	// Receive the first tick. It is expected to fall close to midsecond.
	tt := <-tick.C
	tt = tt.Truncate(maxTimePrecision)
	ch := out

	lastPassed := time.Time{}

	for {
		select {
		case ch <- tt:
			// Tick passed on. Remember its value so that it can be used
			// as a basis for rejecting those ticks from the operating system
			// that would drive tsdbTicker to deliver duplicate ticks.
			lastPassed = tt

			// Prevent re-sending of the tick by temporarily disabling the channel.
			ch = nil

		case tt = <-tick.C:
			if ch != nil {
				// Tick timed out due to slowness of tick consumer.
				// Unfortunate, but cannot do better than dropping it
				// like time.Ticker does.
				statTickerErrors.Add("type=SlowConsumer", 1)
			}
			tt = tt.Truncate(maxTimePrecision)
			if !tt.After(lastPassed) {
				// Passing this tick on would cause tick consumer to
				// produce conflicting data point. Drop it to protect the
				// client. This condition has been observed when CPU-starved.
				// Raising GOMAXPROCS may help, but in cases where this fix
				// is unsuitable (e.g. resource shortage) it is safer to simply drop
				// the tick rather than to cause tick consumer to misbehave.
				statTickerErrors.Add("type=Order", 1)
				continue
			}
			ch = out

		case <-t.stop:
			return
		}
	}
}

func (t *Ticker) Stop() {
	close(t.stop)
}

func Tick(d time.Duration) <-chan time.Time {
	if d < maxTimePrecision {
		log.Panicln("duration too short:", d)
	}
	return NewTicker(d).C
}
