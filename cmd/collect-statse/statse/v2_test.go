// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"testing"
	"time"
)

func messageV2Raw(body string) [][]byte {
	return [][]byte{
		[]byte("2 1234567890000"),
		[]byte(body),
	}
}

func eventV2(metric, tags string, vs ...interface{}) *Event {
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

var testParseEventV2 = []struct {
	in  [][]byte
	out *Event
	err bool
}{
	{
		in:  messageV2Raw(""),
		err: true,
	},
	{
		in:  messageV2Raw("invalid"),
		err: true,
	},
	{
		in:  messageV2Raw(" EVENT"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT "),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT|foo"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT||"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT|foo|"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT|foo||"),
		out: eventV2("foo", ""),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0"),
		out: eventV2("foo", "", Time, float32(0)),
	},
	{
		in:  messageV2Raw("EVENT|foo|op=getFoo|time=0"),
		out: eventV2("foo", "op=getFoo", Time, float32(0)),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0 ttfb=1"),
		out: eventV2("foo", "", Time, float32(0), TTFB, float32(1)),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0 ttfb=1 size=2"),
		out: eventV2("foo", "", Time, float32(0), TTFB, float32(1), Size, float32(2)),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=1.23"),
		out: eventV2("foo", "", Time, float32(1.23)),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=1.2e3"),
		out: eventV2("foo", ""),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0 ttfb="),
		out: eventV2("foo", "", Time, float32(0)),
	},
	{
		in:  messageV2Raw("EVENT|foo|op=getFoo op=getBar|"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0 ttfb=-1"),
		out: eventV2("foo", "", Time, float32(0)),
	},
	{
		in:  messageV2Raw("EVENT|foo||time=0 xxx=0"),
		out: eventV2("foo", "", Time, float32(0)),
	},
	{
		in:  messageV2Raw("EVENT|foo|a=a b=b c=c d=d e=e|"),
		out: eventV2("foo", "a=a b=b c=c d=d e=e"),
	},
	{
		in:  messageV2Raw("EVENT|foo|a=a b=b c=c d=d e=e f=f|"),
		err: true,
	},
	{
		in:  messageV2Raw("EVENT|foo|a|"),
		err: true,
	},
}

func TestParseEventV2(t *testing.T) {
	for i, tt := range testParseEventV2 {
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
			t.Errorf("#%d. bad decode:\nin:  %q\ngot: %v\nwant:%v",
				i, tt.in[1], got, tt.out)
			continue
		}
	}
}

func BenchmarkDecodeV2(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(int64(len(testMessageV2[0]) + len(testMessageV2[1])))

	r := testReaderV2{}
	dec := NewDecoder(r)

	var ev Event
	for i := 0; i < b.N; i++ {
		if err := dec.Decode(&ev); err != nil {
			b.Errorf("decode error: %v", err)
			return
		}
	}
}

type testReaderV2 struct{}

func (testReaderV2) ReadParts() [2][]byte {
	return testMessageV2
}

var testMessageV2 = [2][]byte{
	[]byte("2 1399462546000"),
	[]byte("EVENT|dummy.metric|op=dummyOp|err=false time=1.2 ttfb=0.9"),
}
