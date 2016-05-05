// Copyright 2016 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"hash/fnv"
	"io/ioutil"
	"log"
	"testing"
)

func silenceLogs() {
	log.SetOutput(ioutil.Discard) // Silence logs
}

func TestSeriesHash(t *testing.T) {
	tests := []struct {
		point    Point
		expected int
	}{
		{
			point:    Point{},
			expected: 0,
		},
		{
			point:    Point{metric: []byte("x")},
			expected: 0,
		},
		{
			point:    Point{metric: []byte("x"), tags: []byte(" y=y")},
			expected: 4036,
		},
	}

	for _, tt := range tests {
		cmd := tt.point.put()
		hash := fnv.New32()
		if actual := cmd.SeriesHash(hash); actual != tt.expected {
			t.Errorf("cmd.SeriesHash (%v) => %d, want %d", cmd, actual, tt.expected)
		}
	}
}
