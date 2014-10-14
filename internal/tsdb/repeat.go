// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import "time"

const (
	repeatHeartbeat     = 10 * time.Minute
	repeatMaxAge        = 12 * time.Minute
	repeatCleanupEveryN = 100000
)

type repeatStatus struct {
	time     int64
	value    interface{}
	n        int
	timePrev int64
}

type repeatTester struct {
	bySeries         map[string]*repeatStatus
	cleanupCountdown int
	scratch          [maxLineLength]byte
}

func newRepeatTester() *repeatTester {
	return &repeatTester{
		bySeries:         make(map[string]*repeatStatus),
		cleanupCountdown: repeatCleanupEveryN,
	}
}

// Test evaluates if the given point is a repeat. Test optionally returns one
// preceding data point; this held point must be sent before the given
// point to preserve correctness of line segments.
func (t *repeatTester) Test(point *Point) (isRepeat bool, held cmd) {
	t.cleanup()
	s := point.appendSeries(t.scratch[:0])
	status := t.bySeries[string(s)]
	if status == nil {
		t.bySeries[string(s)] = &repeatStatus{
			time:  point.time,
			value: point.Value(),
			n:     2,
		}
		return
	}
	time, value := point.time, point.Value()
	isRepeat = time > status.time && value == status.value
	needHeartbeat := isRepeat && time-status.time >= int64(repeatHeartbeat)
	switch {
	default:
		// not a repeat
		if status.n == 0 {
			point.time = status.timePrev
			point.setValueValid(status.value)
			held = point.put()
			point.time = time
			point.setValueValid(value)
		}
		*status = repeatStatus{
			time:  time,
			value: value,
			n:     2,
		}
	case needHeartbeat:
		isRepeat = false
		*status = repeatStatus{
			time:  time,
			value: value,
			n:     1,
		}
	case isRepeat:
		if status.n == 2 {
			isRepeat = false
		}
		status.timePrev = time
		if status.n > 0 {
			status.n--
		}
	}
	return
}

func (t *repeatTester) cleanup() {
	t.cleanupCountdown--
	if t.cleanupCountdown > 0 {
		return
	}
	t.cleanupCountdown = repeatCleanupEveryN
	deadline := time.Now().Add(-repeatMaxAge).UnixNano()
	for s, status := range t.bySeries {
		if status.time <= deadline {
			delete(t.bySeries, s)
		}
	}
}
