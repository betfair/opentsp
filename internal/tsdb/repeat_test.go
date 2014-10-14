// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"testing"
	"time"
)

func xpoint(timestamp int64, value interface{}) *Point {
	point, err := NewPoint(time.Unix(timestamp, 0), value, "s")
	if err != nil {
		panic(err)
	}
	return point
}

var repeatTests = []struct {
	in   []*Point
	skip map[int]bool
	held map[int]string
}{
	0: { // trivial case
		in: []*Point{
			xpoint(1000000001, 1),
		},
		skip: map[int]bool{
			0: false,
		},
		held: map[int]string{},
	},
	1: { // repeats once
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 1),
		},
		skip: map[int]bool{
			0: false,
			1: false,
		},
		held: map[int]string{},
	},
	2: { // repeats twice
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 1),
			xpoint(1000000003, 1),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: true,
		},
		held: map[int]string{},
	},
	3: { // updates
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 2),
		},
		skip: map[int]bool{
			0: false,
			1: false,
		},
		held: map[int]string{},
	},
	4: { // repeats once and updates
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 1),
			xpoint(1000000003, 2),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
		},
		held: map[int]string{},
	},
	5: { // repeats twice and updates
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 1),
			xpoint(1000000003, 1),
			xpoint(1000000004, 2),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: true,
			3: false,
		},
		held: map[int]string{
			3: "put s 1000000003 1\n",
		},
	},
	6: { // repeats 3 times and updates
		in: []*Point{
			0: xpoint(1000000001, 1),
			1: xpoint(1000000002, 1),
			2: xpoint(1000000003, 1),
			3: xpoint(1000000004, 1),
			4: xpoint(1000000005, 2),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: true,
			3: true,
			4: false,
		},
		held: map[int]string{
			4: "put s 1000000004 1\n",
		},
	},
	7: { // updates, repeats 3 times, updates
		in: []*Point{
			0: xpoint(1000000000, 1),
			1: xpoint(1000000001, 1000),
			2: xpoint(1000000002, 1000),
			3: xpoint(1000000003, 1000),
			4: xpoint(1000000004, 1000),
			5: xpoint(1000000005, 1000000),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
			3: true,
			4: true,
			5: false,
		},
		held: map[int]string{
			5: "put s 1000000004 1000\n",
		},
	},
	8: { // updates and causes 1 heartbeat
		in: []*Point{
			0: xpoint(1000000000, 1),
			1: xpoint(1000000001, 1000),
			2: xpoint(1000000002, 1000),
			3: xpoint(1000000600, 1000),
			4: xpoint(1000000601, 1000),
			5: xpoint(1000000602, 1000),
			6: xpoint(1000000603, 1000),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
			3: true,
			4: false,
			5: true,
			6: true,
		},
		held: map[int]string{},
	},
	9: { // repeats but updates at exactly the time a heartbeat would be injected
		in: []*Point{
			0: xpoint(1000000000, 1),
			1: xpoint(1000000001, 1000),
			2: xpoint(1000000002, 1000),
			3: xpoint(1000000600, 1000),
			4: xpoint(1000000601, 1000000),
			5: xpoint(1000000602, 1000000),
			6: xpoint(1000000603, 1000000),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
			3: true,
			4: false,
			5: false,
			6: true,
		},
		held: map[int]string{
			4: "put s 1000000600 1000\n",
		},
	},
	10: { // repeats but updates immediately before a heartbeat would be injected
		in: []*Point{
			0: xpoint(1000000000, 1),
			1: xpoint(1000000001, 1000),
			2: xpoint(1000000002, 1000),
			3: xpoint(1000000003, 1000),
			4: xpoint(1000000600, 1000000),
			5: xpoint(1000000601, 1000000),
			6: xpoint(1000000602, 1000000),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
			3: true,
			4: false,
			5: false,
			6: true,
		},
		held: map[int]string{
			4: "put s 1000000003 1000\n",
		},
	},
	11: { // updates, causes 2 heartbeats, and updates again
		in: []*Point{
			0:  xpoint(1000000000, 1),
			1:  xpoint(1000000001, 1000),
			2:  xpoint(1000000002, 1000),
			3:  xpoint(1000000003, 1000),
			4:  xpoint(1000000004, 1000),
			5:  xpoint(1000000601, 1000),
			6:  xpoint(1000000602, 1000),
			7:  xpoint(1000000603, 1000),
			8:  xpoint(1000000604, 1000),
			9:  xpoint(1000001201, 1000),
			10: xpoint(1000001202, 1000),
			11: xpoint(1000001203, 1000),
			12: xpoint(1000001204, 1000),
			13: xpoint(1000001205, 1000000),
		},
		skip: map[int]bool{
			0:  false,
			1:  false,
			2:  false,
			3:  true,
			4:  true,
			5:  false,
			6:  true,
			7:  true,
			8:  true,
			9:  false,
			10: true,
			11: true,
			12: true,
			13: false,
		},
		held: map[int]string{
			13: "put s 1000001204 1000\n",
		},
	},
	12: { // repeats but changes data type
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000002, 1),
			xpoint(1000000003, 1.0),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
		},
		held: map[int]string{},
	},
	13: { // time conflict, same value
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000001, 1),
		},
		skip: map[int]bool{},
		held: map[int]string{},
	},
	14: { // time conflict, different value
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000001, 2),
		},
		skip: map[int]bool{},
		held: map[int]string{},
	},
	15: { // order error
		in: []*Point{
			xpoint(1000000001, 1),
			xpoint(1000000000, 2),
			xpoint(1000000002, 1),
		},
		skip: map[int]bool{},
		held: map[int]string{},
	},
	16: { // repeats but updates immediately after a heartbeat has been injected
		in: []*Point{
			0: xpoint(1000000000, 1),
			1: xpoint(1000000001, 1000),
			2: xpoint(1000000002, 1000),
			3: xpoint(1000000003, 1000),
			4: xpoint(1000000601, 1000),
			5: xpoint(1000000602, 1000000),
		},
		skip: map[int]bool{
			0: false,
			1: false,
			2: false,
			3: true,
			4: false,
			5: false,
		},
		held: map[int]string{},
	},
}

func TestRepeatTester(t *testing.T) {
	for i, tt := range repeatTests {
		tester := newRepeatTester()
		for j, point := range tt.in {
			skip, held := tester.Test(point)
			if len(held) != len(tt.held[j]) {
				t.Errorf("#%d. invalid hold status\n", i)
				t.Errorf("point: %v\n", point)
				t.Errorf("got:   %q\n", held)
				t.Errorf("want:  %q\n", tt.held[j])
				continue
			}
			got := string(held)
			want := tt.held[j]
			if got != want {
				t.Errorf("#%d. invalid hold status for point %d\ngot:  %q\nwant: %q", i, j, got, want)
			}
			if got, want := skip, tt.skip[j]; got != want {
				t.Errorf("#%d. invalid skip status for point %d\ngot:  %v\nwant: %v", i, j, got, want)
			}
		}
	}
}

func TestRepeatTesterTags(t *testing.T) {
	tester := newRepeatTester()
	t0, t1 := time.Unix(0, 0), time.Unix(1, 0)
	a, _ := NewPoint(t0, 0, "foo", "host", "a")
	b, _ := NewPoint(t1, 0, "foo", "host", "b")
	tester.Test(a)
	skip, _ := tester.Test(b)
	if skip {
		t.Errorf("tags difference ignored")
	}
}

func TestRepeatTesterCleanupIsDelayed(t *testing.T) {
	tester := newRepeatTester()
	tester.cleanupCountdown = 1
	t0 := time.Unix(0, 0)
	point, _ := NewPoint(t0.Add(0*time.Second), 0, "foo")
	tester.Test(point)
	if got := len(tester.bySeries); got != 1 {
		t.Errorf("premature cleanup: want 1 entry, got %d", got)
	}
}

func TestRepeatTesterCleanup(t *testing.T) {
	tester := newRepeatTester()
	tester.cleanupCountdown = 2
	now := time.Now()
	point0, _ := NewPoint(now.Add(-repeatMaxAge-1*time.Second), 0, "foo")
	point1, _ := NewPoint(now, 0, "bar")
	tester.Test(point0)
	tester.Test(point1)
	if got := len(tester.bySeries); got != 1 {
		t.Errorf("no cleanup: want 1 entry, got %d", got)
	}
	if tester.bySeries["foo"] != nil {
		t.Errorf("wrong point cleaned up")
	}
}

func BenchmarkRepeatTester(b *testing.B) {
	b.ReportAllocs()
	tester := newRepeatTester()
	t0 := time.Unix(0, 0)
	point0, _ := NewPoint(t0.Add(0*time.Second), 0, "foo", "host", "hhhh", "cluster", "ccccccc")
	point1, _ := NewPoint(t0.Add(1*time.Second), 0, "foo", "host", "hhhh", "cluster", "ccccccc")
	point2, _ := NewPoint(t0.Add(2*time.Second), 0, "foo", "host", "hhhh", "cluster", "ccccccc")
	point3, _ := NewPoint(t0.Add(3*time.Second), 1, "foo", "host", "hhhh", "cluster", "ccccccc")
	for i := 0; i < b.N; i++ {
		tester.Test(point0)
		tester.Test(point1)
		tester.Test(point2)
		tester.Test(point3)
	}
}
