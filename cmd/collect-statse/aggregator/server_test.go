// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

func BenchmarkServer(b *testing.B) {
	const (
		nmetrics      = 1000
		npermutations = 50
		nhosts        = 20
		nstatistics   = 10000
	)

	var (
		event      = statse.Event{}
		metric     []string
		statistics [][]statse.Statistic
	)
	for metricID := 0; metricID < nmetrics; metricID++ {
		metric = append(metric, fmt.Sprintf("metric%d", metricID))
	}

	for statisticID := 0; statisticID < nstatistics; statisticID++ {
		statistics = append(statistics, []statse.Statistic{
			{Key: statse.Time, Value: float32(rand.NormFloat64()*10 + 12.1)},
			{Key: statse.TTFB, Value: float32(rand.NormFloat64()*10 + 10.1)},
			{Key: statse.Size, Value: float32(rand.NormFloat64()*5000 + 12345)},
		})
	}

	client := make([]testServerConn, 10)

NextClient:
	for i := 0; i < len(client); i++ {
		enc := gob.NewEncoder(&client[i].buf)

		metricID := 0
		statisticsID := 0

		for i := 0; ; {
			event.Metric = metric[metricID]
			metricID = (metricID + 1) % len(metric)

			for j := 0; j < npermutations; j++ {
				event.Tags = fmt.Sprintf("tags_permutation=%d", j)

				event.Statistics = statistics[statisticsID]
				statisticsID = (statisticsID + 1) % len(statistics)

				enc.Encode(&event)
				if i++; i >= b.N/len(client) {
					continue NextClient
				}
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	var totalBytes int64
	for _, conn := range client {
		totalBytes += int64(conn.buf.Len())
	}
	b.SetBytes(totalBytes / int64(b.N))

	// Skip the benchmark because the initialisation above is too heavy for use with
	// the -cpuprofile flag.
	// https://code.google.com/p/go/issues/detail?id=8213
	b.Skip()
	binDst, _ := os.Create("/tmp/cpu.test")
	binSrc, _ := os.Open(os.Args[0])
	io.Copy(binDst, binSrc)
	cpuOut, _ := os.Create("/tmp/cpu.out")
	pprof.StartCPUProfile(cpuOut)
	defer pprof.StopCPUProfile()

	store := newStore()
	done := make(chan bool)
	for _, conn := range client {
		conn := conn
		serverConn := &serverConn{&conn, store}
		go func() {
			serverConn.loop()
			done <- true
		}()
	}
	for i := 0; i < len(client); i++ {
		<-done
	}
	job := snapshotJob{
		Time:  time.Unix(0, 0),
		Store: store,
	}
	job.do()
}

type testServerConn struct {
	net.Conn
	buf bytes.Buffer
}

func (c *testServerConn) Read(b []byte) (int, error) {
	return c.buf.Read(b)
}

func (c *testServerConn) Close() error {
	return nil
}
