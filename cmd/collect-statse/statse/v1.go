// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"bytes"
	"fmt"
)

func (ev *Event) parseV1(s []byte) error {
	// Field 1: class name.
	if !bytes.HasPrefix(s, []byte("EVENT ")) {
		return fmt.Errorf("class value invalid")
	}
	s = s[len("EVENT "):]

	// Field 2: metric name.
	i := bytes.IndexByte(s, ' ')
	if i == -1 {
		return fmt.Errorf("missing metric field count")
	}
	ev.Metric = string(s[:i])
	if ev.Metric == "" {
		return fmt.Errorf("invalid metric value: empty string")
	}
	s = s[i+1:]

	// Field 3: value of an implied time statistic.
	i = bytes.IndexByte(s, ' ')
	if i == -1 {
		ev.parseStatistic(Time, s)
		return nil
	}
	ev.parseStatistic(Time, s[:i])
	s = s[i+1:]

	var tagsBuf [64]byte
	tags := tagsBuf[:0]

	// Field 4: value of an implied "op" tag, or a named statistic.
	var v []byte
	if i := bytes.IndexByte(s, ' '); i >= 0 {
		v = s[:i]
		s = s[i+1:]
	} else {
		v = s
		s = s[:0]
	}
	if i := bytes.IndexByte(v, '='); i >= 0 {
		ev.parseStatistics(v, true)
	} else if len(v) > 0 {
		tags = append(tags, []byte("op=")...)
		tags = append(tags, v...)
		ev.Tags = string(tags)
	}

	// Field 5+: named statistics.
	ev.parseStatistics(s, true)

	return nil
}
