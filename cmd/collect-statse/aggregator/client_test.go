// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"bufio"
	"encoding/gob"
	"io/ioutil"
	"testing"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

func BenchmarkClient(b *testing.B) {
	b.ReportAllocs()

	client := &Client{dial: testDial}

	event := &statse.Event{
		Time:   time.Now(),
		Metric: "foo.bar.BazMetric",
		Tags:   "op=getBar host=foo.example.com",
		Statistics: []statse.Statistic{
			{Key: statse.Time, Value: 12.3},
			{Key: statse.TTFB, Value: 12.0},
			{Key: statse.Size, Value: 1234},
		},
	}
	b.SetBytes(int64(16 + len(event.Metric) + len(event.Tags) + 3*(4+4)))

	for i := 0; i < b.N; i++ {
		client.Send(event)
	}
}

func testDial(_ string) *clientConn {
	w := bufio.NewWriter(ioutil.Discard)
	return &clientConn{
		Encoder: gob.NewEncoder(w),
		Writer:  w,
	}
}
