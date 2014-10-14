// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"reflect"
	"testing"
	"time"
)

func messageV1Raw(body string) [][]byte {
	return [][]byte{
		[]byte("1 1234567890000"),
		[]byte(body),
	}
}

func eventV1(metric, tags string, vs ...interface{}) *Event {
	ev := &Event{
		Time:   time.Unix(1234567890, 0),
		Metric: metric,
		Tags:   tags,
	}
	ev.Statistics = ev.statisticsBuf[:0]
	for i := 0; i < len(vs); i += 2 {
		ev.Statistics = append(ev.Statistics, Statistic{
			Key:   vs[i].(Key),
			Value: vs[i+1].(float32),
		})
	}
	return ev
}

var testParseEventV1 = []struct {
	in  [][]byte
	out *Event
	err bool
}{
	{
		in:  messageV1Raw(""),
		err: true,
	},
	{
		in:  messageV1Raw("invalid"),
		err: true,
	},
	{
		in:  messageV1Raw(" EVENT"),
		err: true,
	},
	{
		in:  messageV1Raw("EVENT"),
		err: true,
	},
	{
		in:  messageV1Raw("EVENT "),
		err: true,
	},
	{
		in:  messageV1Raw("EVENT foo"),
		err: true,
	},
	{
		in:  messageV1Raw("EVENT foo 0"),
		out: eventV1("foo", "", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT  foo  0"),
		err: true,
	},
	{
		in:  messageV1Raw("EVENT foo 0 getFoo"),
		out: eventV1("foo", "op=getFoo", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 time=0"),
		out: eventV1("foo", "", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 time=1"),
		out: eventV1("foo", "", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 ttfb=1"),
		out: eventV1("foo", "", Time, float32(0), TTFB, float32(1)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 ttfb=1 size=2"),
		out: eventV1("foo", "", Time, float32(0), TTFB, float32(1), Size, float32(2)),
	},
	{
		in:  messageV1Raw("EVENT foo 1.23"),
		out: eventV1("foo", "", Time, float32(1.23)),
	},
	{
		in:  messageV1Raw("EVENT foo 1.2e3"),
		out: eventV1("foo", ""),
	},
	{
		in:  messageV1Raw("EVENT foo 0 ttfb="),
		out: eventV1("foo", "", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 ttfb=-1"),
		out: eventV1("foo", "", Time, float32(0)),
	},
	{
		in:  messageV1Raw("EVENT foo 0 xxx=0"),
		out: eventV1("foo", "", Time, float32(0)),
	},
}

func TestParseEventV1(t *testing.T) {
	for i, tt := range testParseEventV1 {
		got := new(Event)
		if err := got.parse(tt.in[0], tt.in[1]); err != nil {
			if !tt.err {
				t.Errorf("#%d. unexpected error: %v\nin:%q", i, err, tt.in[1])
			}
			continue
		}
		if tt.err {
			t.Errorf("#%d. unexpected success\nin: %q\ngot:%v", i, tt.in[1], got)
			continue
		}
		if !got.equal(tt.out) {
			t.Errorf("#%d. bad decode:\ngot: %v\nwant:%v", i, got, tt.out)
			continue
		}
	}
}

func (a *Event) equal(b *Event) bool {
	if a.Metric != b.Metric {
		return false
	}
	if a.Tags != b.Tags {
		return false
	}
	if !reflect.DeepEqual(a.Statistics, b.Statistics) {
		return false
	}
	return true
}

func BenchmarkDecodeV1(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(testMessageV1[0]) + len(testMessageV1[1])))

	r := testReaderV1{}
	dec := NewDecoder(r)

	var ev Event
	for i := 0; i < b.N; i++ {
		if err := dec.Decode(&ev); err != nil {
			b.Errorf("decode error: %v", err)
			return
		}
	}
}

type testReaderV1 struct{}

func (testReaderV1) ReadParts() [2][]byte {
	return testMessageV1
}

var testMessageV1 = [2][]byte{
	[]byte("1 1399462546000"),
	[]byte("EVENT dummy.metric 1.2 dummyOp err=false ttfb=0.9"),
}
