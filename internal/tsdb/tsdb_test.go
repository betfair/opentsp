// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type point struct {
	time   time.Time
	value  interface{}
	metric string
	tags   []string
}

func (p point) String() string {
	return fmt.Sprintf("{time:%d value:%T(%#v) metric:%q tags:%q}",
		p.time.UnixNano(), p.value, p.value, p.metric, p.tags)
}

var t0 = time.Unix(0, 0)

var testNewPoint = []struct {
	in, out point
	err     string
}{
	{
		in:  point{time.Time{}, 1, "m", nil},
		err: "time: zero struct",
	},
	{
		in:  point{t0, nil, "m", nil},
		err: "value: nil",
	},
	{
		in:  point{t0, 1, "", nil},
		err: "metric: empty string",
	},
	{
		in:  point{time.Now(), 1, "m", []string{"k"}},
		err: "truncated tags list",
	},
	{
		in:  point{t0, 1, "m", []string{"", "foo"}},
		err: "empty string, in tag ",
	},
	{
		in:  point{t0, 1, "m", []string{"host", ""}},
		err: "empty string, in tag ",
	},
	{
		in:  point{t0, "1e1", "m", nil},
		err: "invalid value.*invalid syntax",
	},
	{
		in:  point{t0, "1.0e+0", "m", nil},
		out: point{t0, float32(1), "m", nil},
	},
	{
		in:  point{t0, "1.", "m", nil},
		out: point{t0, float32(1), "m", nil},
	},
	{
		in:  point{t0, math.NaN(), "m", nil},
		err: "value: NaN",
	},
	{
		in:  point{t0, math.Inf(+1), "m", nil},
		err: "value: Inf",
	},
	{
		in:  point{t0, math.Inf(-1), "m", nil},
		err: "value: Inf",
	},
	{point{t0, "1", "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, "1.0", "m", nil}, point{t0, float32(1), "m", nil}, ""},
	{point{t0, "-1", "m", nil}, point{t0, int64(-1), "m", nil}, ""},
	{point{t0, "-1.0", "m", nil}, point{t0, float32(-1), "m", nil}, ""},
	{point{t0, "1", "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, int(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, int8(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, int16(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, int32(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, int64(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, uint(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, uint8(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, uint16(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, uint32(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, uint64(1), "m", nil}, point{t0, int64(1), "m", nil}, ""},
	{point{t0, float32(1), "m", nil}, point{t0, float32(1), "m", nil}, ""},
	{point{t0, float64(1), "m", nil}, point{t0, float32(1), "m", nil}, ""},
	{
		in:  point{t0, 1, "M", nil},
		out: point{t0, int64(1), "M", nil},
	},
	{
		in:  point{t0, 1, "m", []string{"host", "H"}},
		out: point{t0, int64(1), "m", []string{"host", "H"}},
	},
	{
		in:  point{t0, 1, "a b", nil},
		err: "metric: invalid character",
	},
	{
		in:  point{t0, 1, "m", []string{"host", "a b"}},
		err: "truncated tags list",
	},
	{in: point{t0, 1, "\n", nil}, err: "metric: invalid character"},
	{in: point{t0, 1, "┠", nil}, err: "metric: invalid character"},
	{in: point{t0, 1, "m", []string{"host", "\n"}}, err: "invalid character"},
	{in: point{t0, 1, "m", []string{"host", "\t"}}, err: "invalid character"},
	{in: point{t0, 1, "m0", nil}, out: point{t0, int64(1), "m0", nil}},
	{in: point{t0, 1, "m_", nil}, out: point{t0, int64(1), "m_", nil}},
	{in: point{t0, 1, "m/", nil}, out: point{t0, int64(1), "m/", nil}},
	{in: point{t0, 1, "m.", nil}, out: point{t0, int64(1), "m.", nil}},
	{in: point{t0, 1, "ą", nil}, out: point{t0, int64(1), "ą", nil}},
	{in: point{t0, 1, "m", []string{"host", "m0"}}, out: point{t0, int64(1), "m", []string{"host", "m0"}}},
	{in: point{t0, 1, "m", []string{"host", "m_"}}, out: point{t0, int64(1), "m", []string{"host", "m_"}}},
	{in: point{t0, 1, "m", []string{"host", "m/"}}, out: point{t0, int64(1), "m", []string{"host", "m/"}}},
	{in: point{t0, 1, "m", []string{"host", "m."}}, out: point{t0, int64(1), "m", []string{"host", "m."}}},
	{in: point{t0, 1, "m", []string{"host", "ą"}}, out: point{t0, int64(1), "m", []string{"host", "ą"}}},
	{
		in: point{t0, 1, "m", []string{
			"a", "a",
			"a", "b",
		}},
		err: "duplicate tag key",
	},
	{
		in: point{t0, 1, "m", []string{
			"a", "a",
			"a", "a",
		}},
		err: "duplicate tag key",
	},
	{
		in: point{t0, 1, "m", []string{
			"a", "1",
			"b", "2",
			"c", "3",
			"d", "4",
			"e", "5",
			"f", "6",
			"g", "7",
			"h", "8",
			"i", "9",
		}},
		err: "too many tags",
	},
}

func TestNewPoint(t *testing.T) {
	for _, tt := range testNewPoint {
		p, err := NewPoint(tt.in.time, tt.in.value, tt.in.metric, tt.in.tags...)
		if err != nil {
			switch tt.err {
			case "":
				t.Errorf("NewPoint(%v): unexpected error: %v", tt.in, err)
			default:
				re := regexp.MustCompile(tt.err)
				if !re.MatchString(err.Error()) {
					t.Errorf("NewPoint(%v): got error: %v, want error: %v", tt.in, err, tt.err)
				}
			}
			continue
		}
		if tt.err != "" {
			t.Errorf("NewPoint(%v): unexpected success, want error: %v", tt.in, tt.err)
			continue
		}
		want := tt.in
		if tt.out.value != nil {
			want = tt.out
		}
		if p.Time() != want.time {
			t.Errorf("NewPoint(%v): invalid Time, got %v, want %v", tt.in, p.Time(), want.time)
		}
		if p.Value() != want.value {
			t.Errorf("NewPoint(%v): invalid Value, got %T(%v), want %T(%v)", tt.in,
				p.Value(), p.Value(), want.value, want.value)
		}
		if string(p.Metric()) != want.metric {
			t.Errorf("NewPoint(%v): invalid Metric, got %q, want %q", tt.in,
				p.Metric(), want.metric)
		}
		if !reflect.DeepEqual(p.Tags(), want.tags) {
			t.Errorf("NewPoint(%v): invalid Tags\ngot:   %#v\nwant: %#v", tt.in,
				p.Tags(), want.tags)
		}
	}
}

func BenchmarkNewPoint(b *testing.B) {
	now := time.Now()
	b.ReportAllocs()
	b.SetBytes(80)
	for i := 0; i < b.N; i++ {
		NewPoint(now, "123.456", "os.cpu.seconds",
			"type", "Idle",
			"cluster", "foo",
			"host", "foo.example.com",
		)
	}
}
