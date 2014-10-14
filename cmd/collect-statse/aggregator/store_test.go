// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"reflect"
	"testing"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

var testStoreWrite = []struct {
	in  []*statse.Event
	out map[key]*entry
}{
	{ // initial state
		in:  []*statse.Event{},
		out: map[key]*entry{},
	},
	{ // error: missing host tag
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1.23},
				},
			},
		},
		out: map[key]*entry{},
	},
	{ // successful event: bump success counter and buffer the statistic.
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=bar",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1.23},
				},
			},
		},
		out: map[key]*entry{
			key{Host: "bar", Metric: "foo"}: {
				CountOkay: 1,
				Buffer: [statse.MaxKeys]buffer{
					statse.Time: {1.23},
				},
			},
		},
	},
	{ // erroneous event: bump error counter but don't buffer the statistic.
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=bar",
				Error:  true,
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1},
				},
			},
		},
		out: map[key]*entry{
			key{Host: "bar", Metric: "foo"}: {
				CountError: 1,
			},
		},
	},
	{ // successful event x 2
		in: []*statse.Event{
			{
				Metric: "foo",
				Tags:   "host=bar",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 1.23},
				},
			},
			{
				Metric: "foo",
				Tags:   "host=bar",
				Statistics: []statse.Statistic{
					{Key: statse.Time, Value: 4.56},
				},
			},
		},
		out: map[key]*entry{
			key{Host: "bar", Metric: "foo"}: {
				CountOkay: 2,
				Buffer: [statse.MaxKeys]buffer{
					statse.Time: {1.23, 4.56},
				},
			},
		},
	},
}

func TestStoreWrite(t *testing.T) {
	for i, tt := range testStoreWrite {
		store := newStore()
		store.Write(tt.in...)
		store.Do(func(key key, entry *entry) {
			entry.entryMetadata = entryMetadata{}
		})
		if !reflect.DeepEqual(store.m, tt.out) {
			t.Errorf("#%d. invalid write\nin: %v\ngot: %+v\nwant:%+v",
				i, tt.in, store.m, tt.out)
		}
	}
}

func TestStoreBufferLimit(t *testing.T) {
	store := newStore()
	for i := 0; i < MaxBufferLen+1; i++ {
		store.Write(&statse.Event{
			Metric: "foo",
			Tags:   "host=bar",
			Statistics: []statse.Statistic{
				{
					Key:   statse.Time,
					Value: 1.23,
				},
			},
		})
	}
	entry := store.m[key{Host: "bar", Metric: "foo"}]
	buf := entry.Buffer[statse.Time]
	if len(buf) != MaxBufferLen {
		t.Errorf("buffer length not respected: got %d, want %d",
			len(buf), MaxBufferLen)
	}
}

func TestStoreCleanup(t *testing.T) {
	store := newStore()
	store.Write(&statse.Event{
		Metric: "foo",
		Tags:   "host=a",
	})
	if len(store.m) != 1 {
		t.Fatalf("missing entry")
	}
	time.Sleep(10 * time.Millisecond)
	job := &cleanupJob{store, 0}
	job.do()
	if len(store.m) != 0 {
		t.Fatalf("cleanup fail")
	}
}

var testParseHostTag = []struct {
	in, out, host string
	err           bool
}{
	{
		in:  "",
		err: true,
	},
	{
		in:  "a=a",
		err: true,
	},
	{
		in:  "xhost=a",
		err: true,
	},
	{
		in:   "host=foo",
		out:  "",
		host: "foo",
	},
	{
		in:   "a=a host=foo b=b",
		out:  "a=a b=b",
		host: "foo",
	},
	{
		in:   "a=a host=foo",
		out:  "a=a",
		host: "foo",
	},
	{
		in:  "a=a host=",
		err: true,
	},
}

func TestParseHostTag(t *testing.T) {
	for i, tt := range testParseHostTag {
		event := &statse.Event{Tags: tt.in}
		host, err := parseHostTag(event)
		if err != nil {
			if !tt.err {
				t.Errorf("#%d. unexpected error: %v", i, err)
			}
			continue
		}
		if tt.err {
			t.Errorf("#%d. unexpected success", i)
			continue
		}
		if event.Tags != tt.out {
			t.Errorf("#%d. invalid tags\nin:  %q\ngot: %q\nwant:%q", i, tt.in, event.Tags, tt.out)
		}
		if host != tt.host {
			t.Errorf("#%d. invalid host\nin:  %q\ngot: %q\nwant:%q", i, tt.in, host, tt.host)
		}
	}
}
