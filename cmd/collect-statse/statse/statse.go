// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package statse implements the Statse protocol
package statse

import (
	"bytes"
	"fmt"
	"log"
	"time"
)

var Debug *log.Logger

const (
	// MaxTagsPerEvent limits the number of event tags.
	MaxTagsPerEvent = 5

	// MaxKeys is the count of available Keys.
	MaxKeys = len(keyText)
)

// Defined statistics.
const (
	Time Key = iota
	TTFB
	Size
)

type Key int

func (k Key) String() string {
	return keyText[k]
}

var keyText = [...]string{
	Time: "time",
	TTFB: "ttfb",
	Size: "size",
}

type Statistic struct {
	Key   Key
	Value float32
	init  bool
}

// Event represents performance measurements for an application-specific event.
type Event struct {
	Time          time.Time
	Metric        string
	Tags          string
	Error         bool
	Statistics    []Statistic
	statisticsBuf [MaxKeys]Statistic
}

func (e *Event) String() string {
	buf := new(bytes.Buffer)
	timeMillis := e.Time.UnixNano() / 1e6
	id := e.Metric
	if e.Tags != "" {
		id += " " + e.Tags
	}
	fmt.Fprintf(buf, "Event{@%d %q", timeMillis, id)
	if e.Error {
		fmt.Fprintf(buf, " Error:true")
	}
	for _, stat := range e.Statistics {
		fmt.Fprintf(buf, " %s:%f", stat.Key, stat.Value)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}
