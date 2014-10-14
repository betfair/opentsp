// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package aggregator creates data points based on statse messages.
package aggregator

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"opentsp.org/cmd/collect-statse/statse"

	"opentsp.org/internal/tsdb"
)

type snapshotJob struct {
	Time   time.Time
	Store  *store
	Output []*tsdb.Point
}

func (job *snapshotJob) do() {
	byCluster := make(map[key]*entry)

	// Emit byhost stats.
	job.Store.Do(func(key key, host *entry) {
		// Update the synthetic cluster entry using the host entry.
		clusterKey := key
		clusterKey.Host = ""
		cluster := byCluster[clusterKey]
		if cluster == nil {
			cluster = newEntry()
			byCluster[clusterKey] = cluster
		}
		cluster.CountError += host.CountError
		cluster.CountOkay += host.CountOkay
		for key, values := range host.Buffer {
			for _, v := range values {
				cluster.Buffer[key].Append(v)
			}
		}

		// Process the host entry.
		for _, stat := range calc(host) {
			job.emitf(stat.Value, "%s.byhost.%s %s host=%s",
				key.Metric, stat.Name, key.Tags, key.Host)
			for i := range host.Buffer {
				host.Buffer[i].Reset()
			}
		}
	})

	// Emit bycluster stats.
	for key, cluster := range byCluster {
		for _, stat := range calc(cluster) {
			job.emitf(stat.Value, "%s.%s %s host=NA", key.Metric,
				stat.Name, key.Tags)
		}
	}
}

// emitf emits a data point.
func (job *snapshotJob) emitf(value interface{}, format string, arg ...interface{}) {
	series := fmt.Sprintf(format, arg...)
	id := strings.Fields(strings.Replace(series, "=", " ", -1))
	point, err := tsdb.NewPoint(job.Time, value, id[0], id[1:]...)
	if err != nil {
		log.Printf("aggregator: %v", err)
		return
	}
	job.Output = append(job.Output, point)
}

// calc calculates stats based on the provided entry.
func calc(entry *entry) []stat {
	s := make([]stat, 0, 2+statse.MaxKeys*5)
	s = append(s, stat{"count error=false", entry.CountOkay})
	s = append(s, stat{"count error=true", entry.CountError})
	for i, values := range entry.Buffer {
		if values == nil { // Never updated.
			continue
		}
		if len(values) == 0 { // Updated but not in this cycle.
			// Put in dummy zero value.
			values = []float32{0}
		}
		sort.Sort(ascending(values))
		statseKey := statse.Key(i).String()
		s = append(s, stat{statseKey + ".min", min(values...)})
		s = append(s, stat{statseKey + ".avg", avg(values...)})
		s = append(s, stat{statseKey + ".p95", p95(values...)})
		s = append(s, stat{statseKey + ".p99", p99(values...)})
		s = append(s, stat{statseKey + ".max", max(values...)})
	}
	return s
}

func min(a ...float32) float32 { return a[0] }
func avg(a ...float32) float32 { return sum(a...) / float32(len(a)) }
func p95(a ...float32) float32 { return a[95*len(a)/100] }
func p99(a ...float32) float32 { return a[99*len(a)/100] }
func max(a ...float32) float32 { return a[len(a)-1] }

func sum(a ...float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i]
	}
	return sum
}

// stat represents a statistic based on a buffer.
type stat struct {
	Name  string
	Value interface{}
}

type ascending []float32

func (a ascending) Len() int           { return len(a) }
func (a ascending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ascending) Less(i, j int) bool { return a[i] < a[j] }
