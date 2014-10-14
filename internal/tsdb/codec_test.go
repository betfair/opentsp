// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"io/ioutil"
	"testing"
)

var testMarshalText = []struct {
	in  Point
	out string
	err bool
}{
	{
		in: Point{
			time:     1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
		},
		out: "x 1234567890 1",
	},
	{
		in: Point{
			time:       1234567890 * 1e9,
			valueFloat: float32(1),
			isFloat:    true,
			metric:     []byte("x"),
		},
		out: "x 1234567890 1.0",
	},
	{
		in: Point{
			time:     1234567890 * 1e9,
			valueInt: int64(-1),
			metric:   []byte("x"),
		},
		out: "x 1234567890 -1",
	},
	{
		in: Point{
			time:       1234567890 * 1e9,
			valueFloat: float32(-1),
			isFloat:    true,
			metric:     []byte("x"),
		},
		out: "x 1234567890 -1.0",
	},
	{
		in: Point{time: 1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
			tags:     []byte(" y=y z=z"),
		},
		out: "x 1234567890 1 y=y z=z",
	},
}

func TestMarshalText(t *testing.T) {
	for _, tt := range testMarshalText {
		var buf []byte
		p := tt.in
		buf = p.append(buf)
		if tt.err {
			t.Errorf("MarshalText(%v): got success, want error", &tt.in)
			continue
		}
		if string(buf) != tt.out {
			t.Errorf("MarshalText(%v): invalid output", &tt.in)
			t.Errorf("  have=%q", buf)
			t.Errorf("  want=%q", tt.out)
			continue
		}
	}
}

var testUnmarshalText = []struct {
	s   string
	err bool
	p   Point
}{
	{s: "", err: true},
	{s: "x", err: true},
	{s: "x x", err: true},
	{s: "x badt 0", err: true},
	{s: "x 1234567890 badv", err: true},
	{s: "x 1234567890 0 k=", err: true},
	{s: "x 1234567890 0 =v", err: true},
	{s: "x 1234567890 0 =", err: true},
	{s: "x -123456789 1", err: true},
	{
		s: "x 1234567890 1",
		p: Point{
			time:     1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
		},
	},
	{
		s: "x 1234567890 1",
		p: Point{
			time:     1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
		},
	},
	{
		s: "x 1234567890 1 y=y",
		p: Point{time: 1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
			tags:     []byte(" y=y"),
		},
	},
	{
		s:   " x 1234567890 1 y=y",
		err: true,
	},
	{
		s: "x 1234567890 1 y=y  ",
		p: Point{time: 1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
			tags:     []byte(" y=y"),
		},
	},
	{
		s:   "\tx 1234567890 1 y=y",
		err: true,
	},
	{
		s: "x\t1234567890\t1\ty=y\t\t",
		p: Point{time: 1234567890 * 1e9,
			valueInt: int64(1),
			metric:   []byte("x"),
			tags:     []byte(" y=y"),
		},
	},
}

func TestUnmarshalText(t *testing.T) {
	for _, tt := range testUnmarshalText {
		p := Point{}
		err := p.unmarshalText([]byte(tt.s))
		if err != nil {
			if !tt.err {
				t.Errorf("unmarshalText(%q): got error %q, want success", tt.s, err)
			}
			continue
		}
		if tt.err {
			t.Errorf("unmarshalText(%q): got success, want error", tt.s)
			continue
		}
		if !p.Equal(&tt.p) {
			t.Errorf("unmarshalText(%q): invalid point", tt.s)
			t.Errorf("  have=%v", &p)
			t.Errorf("  want=%v", &tt.p)
			continue
		}
	}
}

func BenchmarkUnmarshalText(b *testing.B) {
	point := []byte("xxx.xxx.xxx  1234567890  111  a=a b=b c=c")
	b.SetBytes(int64(len(point)))
	b.ReportAllocs()
	var p Point
	for i := 0; i < b.N; i++ {
		p = Point{}
		p.unmarshalText(point)
	}
}

const pointText = "xxx.xxx.xxx  1234567890  111  a=a b=b c=c\n"

func BenchmarkAppend(b *testing.B) {
	var p Point
	p.unmarshalText([]byte(pointText))
	b.SetBytes(int64(len(pointText)))
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = buf[:0]
		buf = p.append(buf)
	}
}

func BenchmarkEncode(b *testing.B) {
	b.SetBytes(int64(len(pointText)))
	b.ReportAllocs()
	var p Point
	p.unmarshalText([]byte(pointText))
	enc := NewEncoder(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		err := enc.Encode(&p)
		if err != nil {
			b.Fatal(err)
		}
	}
}
