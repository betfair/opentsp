// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package filter

import (
	"testing"
	"time"

	"opentsp.org/internal/tsdb"
)

var varTrue, varFalse = true, false

var testNew = []struct {
	rules []Rule
	err   bool
}{
	{ // a pass rule
		rules: []Rule{
			{
				Set:   []string{},
				Block: &varFalse,
			},
		},
		err: false,
	},
	{ // missing tag value in tag match
		rules: []Rule{
			{
				Match: []string{
					".*",
					"host",
				},
			},
		},
		err: true,
	},
	{ // undefined SetMetric
		rules: []Rule{
			{
				Set: []string{""},
			},
		},
		err: true,
	},
	{ // missing tag value in SetTag
		rules: []Rule{
			{
				Set: []string{
					"",
					"a",
				},
			},
		},
		err: true,
	},
	{ // single submatch-using regex, but no references
		rules: []Rule{
			{
				Match: []string{
					"foo(.*)",
				},
				Set: []string{"bar"},
			},
		},
	},
	{ // multiple submatch-using regexes
		rules: []Rule{
			{
				Match: []string{
					"foo(.*)",
					"host",
					"bar(.*)",
				},
				Set: []string{"${1}"},
			},
		},
		err: true,
	},
	{ // a reference to non-existent submatch
		rules: []Rule{
			{
				Match: []string{
					"foo(.*)",
				},
				Set: []string{"${0}"},
			},
		},
		err: true,
	},
	{ // a reference to non-existent submatch
		rules: []Rule{
			{
				Match: []string{
					"foo(.*)",
				},
				Set: []string{"${2}"},
			},
		},
		err: true,
	},
	{ // using a submatch to set the tag name
		rules: []Rule{
			{
				Match: []string{
					"foo(.*)",
				},
				Set: []string{
					"",
					"${2}",
					"foo",
				},
			},
		},
		err: true,
	},
	{ // mutating and blocking
		rules: []Rule{
			{
				Set:   []string{"foo"},
				Block: &varTrue,
			},
		},
		err: true,
	},
	{ // mutating and blocking
		rules: []Rule{
			{
				Set: []string{
					"",
					"foo",
					"1",
				},
				Block: &varTrue,
			},
		},
		err: true,
	},
	{ // allow mutating and Block=false (which means pass/accept)
		rules: []Rule{
			{
				Set:   []string{"foo"},
				Block: &varFalse,
			},
		},
		err: false,
	},
	{ // empty tag value in SetTag
		rules: []Rule{
			{
				Set: []string{
					"",
					"a",
					"",
				},
			},
		},
		err: true,
	},
	{ // submatch defined but unused
		rules: []Rule{
			{
				Match: []string{"(a|b)"},
				Block: &varTrue,
			},
		},
	},
}

func TestNew(t *testing.T) {
	for i, tt := range testNew {
		_, err := New(tt.rules...)
		if err != nil {
			if !tt.err {
				t.Errorf("#%d. unexpected error: %v for test: %v", i, err, tt)
			}
			continue
		}
		if tt.err {
			t.Errorf("#%d. unexpected success for test: %v", i, tt)
			continue
		}
	}
}

func point(metric string, keyval ...string) *tsdb.Point {
	point, err := tsdb.NewPoint(time.Unix(0, 0), 0, metric, keyval...)
	if err != nil {
		panic(err)
	}
	return point
}

var testEval = []struct {
	in, out *tsdb.Point
	rules   []Rule
	pass    bool
}{
	0: { // block rule
		in: point("foo"),
		rules: []Rule{
			{Block: &varTrue},
		},
		pass: false,
	},
	1: { // metric override
		in: point("foo"),
		rules: []Rule{
			{Set: []string{"bar"}},
		},
		out:  point("bar"),
		pass: true,
	},
	2: { // tag override
		in: point("foo", "op", "getFoo"),
		rules: []Rule{
			{Set: []string{"", "op", "getBar"}},
		},
		out:  point("foo", "op", "getBar"),
		pass: true,
	},
	3: { // tag create
		in: point("foo"),
		rules: []Rule{
			{Set: []string{"", "host", "web01"}},
		},
		out:  point("foo", "host", "web01"),
		pass: true,
	},
	4: { // metric prefix
		in: point("foo"),
		rules: []Rule{
			{
				Match: []string{"(.*)"},
				Set:   []string{"adhoc.${1}"},
			},
		},
		out:  point("adhoc.foo"),
		pass: true,
	},
	5: { // combine multiple metrics into one
		in: point("foo.bar.baz"),
		rules: []Rule{
			{
				Match: []string{`foo\.([^\.]+)\.(.*)`},
				Set: []string{
					"foo.${2}",
					"newtag",
					"${1}",
				},
			},
		},
		out:  point("foo.baz", "newtag", "bar"),
		pass: true,
	},
	6: { // block path-like tag value: tag present with legal value
		in: point("foo", "op", "_some_path"),
		rules: []Rule{
			{
				Match: []string{
					"",
					"op",
					"^/",
				},
				Block: &varTrue,
			},
		},
		pass: true,
	},
	7: { // block path-like tag value: tag present with illegal value
		in: point("foo", "op", "/some/path"),
		rules: []Rule{
			{
				Match: []string{
					"",
					"op",
					"^/",
				},
				Block: &varTrue,
			},
		},
		pass: false,
	},
	8: { // set 2 new tags
		in: point("foo"),
		rules: []Rule{
			{Set: []string{"", "a", "a", "b", "b"}},
		},
		out:  point("foo", "a", "a", "b", "b"),
		pass: true,
	},
	9: { // set 2 tags, 1 new
		in: point("foo", "a", "a", "b", "b"),
		rules: []Rule{
			{Set: []string{"", "c", "c", "b", "B"}},
		},
		out:  point("foo", "c", "c", "b", "B", "a", "a"),
		pass: true,
	},
}

func TestEval(t *testing.T) {
	for i, tt := range testEval {
		filter, err := New(tt.rules...)
		if err != nil {
			t.Errorf("#%d. invalid test: error creating filter: %v", i, err)
			continue
		}
		got := tt.in.Copy()
		pass, err := filter.Eval(got)
		if err != nil {
			panic(err)
		}
		if !pass {
			if tt.pass {
				t.Errorf("#%d. unexpected block", i)
			}
			continue
		}
		if !tt.pass {
			t.Errorf("#%d. unexpected pass", i)
			continue
		}
		if want := tt.out; want != nil {
			if !got.Equal(want) {
				t.Errorf("#%d. invalid edit:\nin:   %v\ngot:  %v\nwant: %v",
					i, tt.in, got, want)
			}
		}
	}
}

func BenchmarkEval(b *testing.B) {
	b.ReportAllocs()
	b.SetBytes(123) // roughly
	f, err := New(
		Rule{Set: []string{"", "host", "test.host"}},
		Rule{Set: []string{"", "c", "test.cluster"}},
		Rule{Match: []string{"(.*)"}, Set: []string{"adhoc.${1}"}},
		Rule{
			Match: []string{"", "path", "^(/[^/]+)/"},
			Set:   []string{"", "path", "${1}"},
		},
		Rule{
			Match: []string{"", "op", "^(/[^/]+)/"},
			Set:   []string{"", "op", "${1}"},
		},
	)
	if err != nil {
		b.Fatalf("benchmark error: error creating filter: %v", err)
	}
	orig := point("foo.bar.baz",
		"path", "/foofoofoo/barbarbar",
		"aaaaa", "AAAAAAAA",
		"bbbbb", "BBBBBBBBBBBB",
	)
	var point tsdb.Point
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		point = *orig
		b.StartTimer()
		pass, err := f.Eval(&point)
		if err != nil {
			b.Fatalf("error: %v", err)
		}
		if !pass {
			b.Errorf("unexpected block")
		}
	}
}
