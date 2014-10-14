// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"opentsp.org/cmd/collect-statse/statse"

	"opentsp.org/internal/tsdb"
)

func point(time time.Time, value interface{}, series string) *tsdb.Point {
	id := strings.Fields(strings.Replace(series, "=", " ", -1))
	point, err := tsdb.NewPoint(time, value, id[0], id[1:]...)
	if err != nil {
		log.Panicf("point: %v, time=%v value=%v series=%q id=%q", err, time, value, series, id)
	}
	return point
}

var testSnapshot = []struct {
	in  []*statse.Event
	out []*tsdb.Point
}{
	{ // 0 events
		in:  []*statse.Event{},
		out: []*tsdb.Point(nil),
	},
	{ // 1 event
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=a",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1},
				},
			},
		},
		out: []*tsdb.Point{
			// Per-host.
			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=false  host=a"),
			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.min  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.avg  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p95  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p99  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.max  host=a"),

			//  Cluster-wide.
			point(time.Unix(0, 0), uint64(1), "foo.count error=false  host=NA"),
			point(time.Unix(0, 0), uint64(0), "foo.count error=true  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.min  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.avg  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.p95  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.p99  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.max  host=NA"),
		},
	},
	{ // 2 events, 1 host
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=a",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1},
				},
			},
			{
				Metric: "foo",
				Tags:   "host=a",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 2},
				},
			},
		},
		out: []*tsdb.Point{
			// Per-host.
			point(time.Unix(0, 0), uint64(2), "foo.byhost.count error=false  host=a"),
			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.min  host=a"),
			point(time.Unix(0, 0), float32(1.5), "foo.byhost.time.avg  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.time.p95  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.time.p99  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.time.max  host=a"),

			//  Cluster-wide.
			point(time.Unix(0, 0), uint64(2), "foo.count error=false  host=NA"),
			point(time.Unix(0, 0), uint64(0), "foo.count error=true  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.min  host=NA"),
			point(time.Unix(0, 0), float32(1.5), "foo.time.avg  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.time.p95  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.time.p99  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.time.max  host=NA"),
		},
	},
	/*
	   BUG: depends on map iteration order.
	   	{ // 2 events, 2 hosts
	   		in: []*statse.Event{
	   			{
	   				Metric: "foo",
	   				Tags:   "host=a",
	   				Statistics: []statse.Statistic{
	   					{Key: statse.Time, Value: 1},
	   				},
	   			},
	   			{
	   				Metric: "foo",
	   				Tags:   "host=b",
	   				Statistics: []statse.Statistic{
	   					{Key: statse.Time, Value: 2},
	   				},
	   			},
	   		},
	   		out: []*tsdb.Point{
	   			// host=a
	   			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=false  host=a"),
	   			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=a"),
	   			point(time.Unix(0, 0), float32(1), "foo.byhost.time.min  host=a"),
	   			point(time.Unix(0, 0), float32(1), "foo.byhost.time.avg  host=a"),
	   			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p95  host=a"),
	   			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p99  host=a"),
	   			point(time.Unix(0, 0), float32(1), "foo.byhost.time.max  host=a"),

	   			// host=b
	   			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=false  host=b"),
	   			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=b"),
	   			point(time.Unix(0, 0), float32(2), "foo.byhost.time.min  host=b"),
	   			point(time.Unix(0, 0), float32(2), "foo.byhost.time.avg  host=b"),
	   			point(time.Unix(0, 0), float32(2), "foo.byhost.time.p95  host=b"),
	   			point(time.Unix(0, 0), float32(2), "foo.byhost.time.p99  host=b"),
	   			point(time.Unix(0, 0), float32(2), "foo.byhost.time.max  host=b"),

	   			//  Cluster-wide.
	   			point(time.Unix(0, 0), uint64(2), "foo.count error=false  host=NA"),
	   			point(time.Unix(0, 0), uint64(0), "foo.count error=true  host=NA"),
	   			point(time.Unix(0, 0), float32(1), "foo.time.min  host=NA"),
	   			point(time.Unix(0, 0), float32(1.5), "foo.time.avg  host=NA"),
	   			point(time.Unix(0, 0), float32(2), "foo.time.p95  host=NA"),
	   			point(time.Unix(0, 0), float32(2), "foo.time.p99  host=NA"),
	   			point(time.Unix(0, 0), float32(2), "foo.time.max  host=NA"),
	   		},
	   	},
	*/
	{ // 1 event, all statistic types
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=a",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1},
					{Key: statse.TTFB, Value: 2},
					{Key: statse.Size, Value: 3},
				},
			},
		},
		out: []*tsdb.Point{
			// byhost count
			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=false  host=a"),
			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=a"),

			// byhost time
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.min  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.avg  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p95  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.p99  host=a"),
			point(time.Unix(0, 0), float32(1), "foo.byhost.time.max  host=a"),

			// byhost ttfb
			point(time.Unix(0, 0), float32(2), "foo.byhost.ttfb.min  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.ttfb.avg  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.ttfb.p95  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.ttfb.p99  host=a"),
			point(time.Unix(0, 0), float32(2), "foo.byhost.ttfb.max  host=a"),

			// byhost size
			point(time.Unix(0, 0), float32(3), "foo.byhost.size.min  host=a"),
			point(time.Unix(0, 0), float32(3), "foo.byhost.size.avg  host=a"),
			point(time.Unix(0, 0), float32(3), "foo.byhost.size.p95  host=a"),
			point(time.Unix(0, 0), float32(3), "foo.byhost.size.p99  host=a"),
			point(time.Unix(0, 0), float32(3), "foo.byhost.size.max  host=a"),

			// cluster count
			point(time.Unix(0, 0), uint64(1), "foo.count error=false  host=NA"),
			point(time.Unix(0, 0), uint64(0), "foo.count error=true  host=NA"),

			// cluster time
			point(time.Unix(0, 0), float32(1), "foo.time.min  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.avg  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.p95  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.p99  host=NA"),
			point(time.Unix(0, 0), float32(1), "foo.time.max  host=NA"),

			// cluster ttfb
			point(time.Unix(0, 0), float32(2), "foo.ttfb.min  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.ttfb.avg  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.ttfb.p95  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.ttfb.p99  host=NA"),
			point(time.Unix(0, 0), float32(2), "foo.ttfb.max  host=NA"),

			// cluster size
			point(time.Unix(0, 0), float32(3), "foo.size.min  host=NA"),
			point(time.Unix(0, 0), float32(3), "foo.size.avg  host=NA"),
			point(time.Unix(0, 0), float32(3), "foo.size.p95  host=NA"),
			point(time.Unix(0, 0), float32(3), "foo.size.p99  host=NA"),
			point(time.Unix(0, 0), float32(3), "foo.size.max  host=NA"),
		},
	},
	{ // an event without statistics
		in: []*statse.Event{
			{
				Metric:     "foo",
				Tags:       "host=a",
				Statistics: []statse.Statistic{},
			},
		},
		out: []*tsdb.Point{
			// byhost count
			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=false  host=a"),
			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=true  host=a"),

			// cluster count
			point(time.Unix(0, 0), uint64(1), "foo.count error=false  host=NA"),
			point(time.Unix(0, 0), uint64(0), "foo.count error=true  host=NA"),
		},
	},
	{ // an error event with statistics
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=a",
				Error:  true,
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1},
					{Key: statse.TTFB, Value: 2},
					{Key: statse.Size, Value: 3},
				},
			},
		},
		out: []*tsdb.Point{
			// byhost count
			point(time.Unix(0, 0), uint64(0), "foo.byhost.count error=false  host=a"),
			point(time.Unix(0, 0), uint64(1), "foo.byhost.count error=true  host=a"),

			// cluster count
			point(time.Unix(0, 0), uint64(0), "foo.count error=false  host=NA"),
			point(time.Unix(0, 0), uint64(1), "foo.count error=true  host=NA"),
		},
	},
}

func TestSnapshot(t *testing.T) {
	for i, tt := range testSnapshot {
		store := newStore()
		store.Write(tt.in...)
		job := snapshotJob{
			Time:  time.Unix(0, 0),
			Store: store,
		}
		job.do()
		if !reflect.DeepEqual(job.Output, tt.out) {
			t.Errorf("#%d. invalid snapshot", i)
			t.Errorf("in:\n")
			for _, event := range tt.in {
				t.Errorf("	%v\n", event)
			}
			t.Errorf("got:\n")
			for _, point := range job.Output {
				t.Errorf("	%v\n", point)
			}
			t.Errorf("want:\n")
			for _, point := range tt.out {
				t.Errorf("	%v\n", point)
			}
		}
		for key, entry := range store.m {
			for _, buf := range entry.Buffer {
				if len(buf) > 0 {
					t.Errorf("events still buffered following a snapshot, key=%+v", key)
				}
			}
		}
	}
}

func TestSnapshotP95(t *testing.T) {
	store := newStore()
	for i := 1; i <= 100; i++ {
		store.Write(&statse.Event{
			Metric: "foo",
			Tags:   "host=a",
			Statistics: []statse.Statistic{
				{Key: statse.Time, Value: float32(i)},
			},
		})
	}
	job := snapshotJob{
		Time:  time.Unix(0, 0),
		Store: store,
	}
	job.do()
	for _, point := range job.Output {
		if string(point.Metric()) != "foo.time.p95" {
			continue
		}
		if got := point.Value().(float32); got != 96 {
			t.Errorf("invalid p95: got %v, want 96, point=%v", got, point)
		}
		return
	}
	t.Errorf("foo.time.p95 not found")
}

func TestSnapshotNoop(t *testing.T) {
	store := newStore()

	// add some
	store.Write(&statse.Event{
		Metric:     "foo",
		Tags:       "host=a",
		Statistics: []statse.Statistic{{Key: statse.Time, Value: 0}},
	})

	// snapshot
	job := snapshotJob{
		Time:  time.Unix(0, 0),
		Store: store,
	}
	job.do()

	// snapshot (noop)
	job = snapshotJob{
		Time:  time.Unix(1, 0),
		Store: store,
	}
	job.do()

	// ok - no crash.
}
