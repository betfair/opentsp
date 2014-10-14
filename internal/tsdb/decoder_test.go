// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var testDecode = []struct {
	in  string
	err string
}{
	{
		in: timeSequence(
			"1000000001",
			"1000000002",
		),
		err: "",
	},
	{
		in: timeSequence(
			"1000000001001",
			"1000000002000",
		),
		err: "",
	},
	{
		in: timeSequence(
			"1000000001",
			"1000000001",
		),
		err: "order error: collision at time 1000000001",
	},
	{
		in: timeSequence(
			"1000000001001",
			"1000000001002",
		),
		err: "order error: collision at time 1000000001",
	},
	{
		in: timeSequence(
			"1000000001001",
			"1000000001999",
		),
		err: "order error: collision at time 1000000001",
	},
	{
		in: timeSequence(
			"1000000001",
			"1000000000",
		),
		err: "order error: got time 1000000000, want at least 1000000002",
	},
	{
		in: timeSequence(
			"1000000001",
			"1000000002",
			"1000000000",
		),
		err: "order error: got time 1000000000, want at least 1000000003",
	},
	{
		in: timeSequence(
			"1000000001",
			"1000086402",
		),
		err: "order error: stepped too far into the future \\(24h0m1s>24h0m0s\\)",
	},
	{
		in: timeSequence(
			"1000000001",
			"1000000001000",
		),
		err: "order error: collision at time 1000000001",
	},
	/*
		{
			in: timeSequence(
				"2147483647",
				"1000000001",
			),
			err: "order error: stepped too far into the future",
		},
	*/
}

func TestDecode(t *testing.T) {
	for _, tt := range testDecode {
		dec := NewDecoder(strings.NewReader(tt.in))
		var err error
		for {
			_, err = dec.Decode()
			if err != nil {
				break
			}
		}
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			if tt.err == "" {
				t.Errorf("Decode(%q): unexpected error:\ngot: %v\nwant success", tt.in, err)
				continue
			}
			re := regexp.MustCompile(tt.err)
			if !re.MatchString(err.Error()) {
				t.Errorf("Decode(%q): invalid error:\ngot: %v\nwant: /%s/", tt.in, err, tt.err)
			}
			continue
		}
		if tt.err != "" {
			t.Errorf("Decode(%q):\ngot success\nwant error: %v", tt.in, tt.err)
			continue
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	b.SetBytes(int64(len(pointText)))
	b.ReportAllocs()
	dec := NewDecoder(&infiniteReader{})
	for i := 0; i < b.N; i++ {
		p, err := dec.Decode()
		if err != nil {
			b.Fatal(err)
		}
		p.Free()
	}
}

type infiniteReader struct {
	time int64
}

func (r *infiniteReader) Read(b []byte) (int, error) {
	r.time++
	b = b[:0]
	b = append(b, "xxx.xxx.xxx  "...)
	b = strconv.AppendInt(b, r.time, 10)
	b = append(b, " 111  a=a b=b c=c\n"...)
	return len(b), nil
}

func timeSequence(times ...string) string {
	var input string
	for _, t := range times {
		input += fmt.Sprintf("foo %s 1 a=a\n", t)
	}
	return input
}
