// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package forwarder

import (
	"bytes"

	"opentsp.org/cmd/collect-statse/statse"
)

// Event is a statse.Event wrapper that adds methods of the filter.Point interface.
type Event struct {
	Statse    statse.Event
	metric    []byte
	metricBuf [128]byte
	tags      []byte
	tagsBuf   [128]byte
}

func (e *Event) String() string {
	ometric := e.Statse.Metric
	otags := e.Statse.Tags
	e.Statse.Metric = string(e.metric)
	e.Statse.Tags = string(e.tags)
	s := e.Statse.String()
	e.Statse.Metric = ometric
	e.Statse.Tags = otags
	return s
}

// Final returns a potentially modified version of the wrapped statse.Event.
func (e *Event) Final() *statse.Event {
	e.Statse.Metric = string(e.metric)
	e.Statse.Tags = string(e.tags)
	return &e.Statse
}

// Reset implements the Reset method of the filter.Point interface.
func (e *Event) Reset() {
	e.metric = append(e.metricBuf[:0], e.Statse.Metric...)
	e.tags = append(e.tagsBuf[:0], e.Statse.Tags...)
}

// Metric implements the Metric method of the filter.Point interface.
func (e *Event) Metric() []byte {
	return e.metric
}

// SetMetric implements the SetMetric method of the filter.Point interface.
func (e *Event) SetMetric(s []byte) error {
	e.metric = append(e.metricBuf[:0], s...)
	return nil
}

// Tag implements the Tag method of the filter.Point interface.
func (e *Event) Tag(key []byte) []byte {
	i := bytes.Index(e.tags, key)
	if i == -1 {
		return nil
	}
	if i > 0 && e.tags[i-1] != ' ' {
		return nil
	}
	if e.tags[i+len(key)] != '=' {
		return nil
	}
	tagv := e.tags[i+len(key)+1:]
	if i = bytes.IndexByte(tagv, ' '); i != -1 {
		tagv = tagv[:i]
	}
	return tagv
}

var (
	sepNone  = []byte("")
	sepSpace = []byte(" ")
)

// SetTags implements the SetTags method of the Event interface.
func (e *Event) SetTags(newval ...[]byte) error {
	buf := make([]byte, 0, 128)
	tags := e.tags

	// Mark all tag keys as undone.
	var doneArray [8 * 2]bool
	done := doneArray[:0]
	for i := 0; i < len(newval); i++ {
		done = append(done, false)
	}

	sep := sepNone
	for len(tags) > 0 {
		buf = append(buf, sep...)
		sep = sepSpace

		// Find next tag.
		var tag []byte
		i := bytes.IndexByte(tags, ' ')
		if i == -1 {
			tag = tags
			tags = tags[:0]
		} else {
			tag = tags[:i]
			tags = tags[i+1:]
		}
		tagk := tag[:bytes.IndexByte(tag, '=')]

		// Check if the tag has a replacement value.
		var tagv []byte
		for i := 0; i < len(newval); i += 2 {
			k := newval[i]
			v := newval[i+1]
			if !bytes.Equal(tagk, k) {
				continue
			}
			tagv = v
			done[i] = true
			break
		}
		if tagv == nil { // No replacement value.
			buf = append(buf, tag...)
			continue
		}
		buf = append(buf, tagk...)
		buf = append(buf, '=')
		buf = append(buf, tagv...)
	}

	// Create new tags if necessary.
	for i := 0; i < len(newval); i += 2 {
		if done[i] {
			continue
		}
		buf = append(buf, sep...)
		sep = sepSpace
		tagk := newval[i]
		tagv := newval[i+1]
		buf = append(buf, tagk...)
		buf = append(buf, '=')
		buf = append(buf, tagv...)
	}

	e.tags = append(e.tagsBuf[:0], buf...)
	return nil
}
