// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"
)

const maxLineLength = 1023 // limit in net.opentsdb.tsd.PipelineFactory

func (p *Point) unmarshalText(buf []byte) error {
	switch {
	default:
		// ok
	case len(buf) == 0:
		return fmt.Errorf("tsdb: invalid point: empty string")
	case len(buf) > maxLineLength:
		return fmt.Errorf("tsdb: invalid point: line too long (%d>%d)", len(buf), maxLineLength)
	}
	p.reset()
	originalBuf := buf
	metric, buf := skipNonSpace(buf)
	if err := p.setMetric(metric); err != nil {
		return fmt.Errorf("tsdb: invalid metric: %v: %q, in %q", err, metric, originalBuf)
	}
	buf = skipSpace(buf)
	timeText, buf := skipNonSpace(buf)
	t, err := parseTime(timeText)
	if err != nil {
		return fmt.Errorf("tsdb: invalid time: %v, in %q", err, originalBuf)
	}
	if err := p.setTime(t); err != nil {
		return fmt.Errorf("tsdb: invalid time: %v, in %q", err, originalBuf)
	}
	buf = skipSpace(buf)
	valueText, buf := skipNonSpace(buf)
	if err := p.setValueBytes(valueText); err != nil {
		return fmt.Errorf("tsdb: invalid value: %v, in %q", err, originalBuf)
	}
	if err := p.setTags(buf, skipNonSpace); err != nil {
		return fmt.Errorf("tsdb: invalid tags: %v, in %q", err, originalBuf)
	}
	return nil
}

// append appends the marshalled version of the point to the end of the provided
// slice. If the slice has insufficient capacity, it is grown using the built-in
// append. In any case, append returns the updated slice.
//
// BUG(masiulaniecj): Millisecond resolution is accepted by Decode and
// NewPoint, but lost by Encode and Put.
func (p *Point) append(b []byte) []byte {
	b = append(b, p.metric...)
	b = append(b, ' ')
	b = appendInt(b, p.time/1e9)
	b = append(b, ' ')
	if p.isFloat {
		b = appendFloat(b, p.valueFloat)
	} else {
		b = appendInt(b, p.valueInt)
	}
	b = append(b, p.tags...)
	return b
}

func (p *Point) appendSeries(buf []byte) []byte {
	buf = append(buf, p.metric...)
	buf = append(buf, p.tags...)
	return buf
}

func appendInt(buf []byte, n int64) []byte {
	tmp := make([]byte, 0, 32)
	neg := false
	if n < 0 {
		neg = true
		n *= -1
	}
	for n >= 10 {
		rem := n % 10
		tmp = append(tmp, '0'+byte(rem))
		n /= 10
	}
	tmp = append(tmp, '0'+byte(n))
	if neg {
		buf = append(buf, '-')
	}
	for i := len(tmp) - 1; i >= 0; i-- {
		buf = append(buf, tmp[i])
	}
	return buf
}

func appendFloat(buf []byte, n float32) []byte {
	tmp := make([]byte, 0, 32)
	tmp = strconv.AppendFloat(tmp, float64(n), 'f', -1, 32)
	if bytes.IndexByte(tmp, '.') == -1 {
		tmp = append(tmp, ".0"...)
	}
	buf = append(buf, tmp...)
	return buf
}

var dot = []byte(".")

func parseTime(b []byte) (time.Time, error) {
	if !(len(b) > 0 && (len(b) <= 10 || len(b) == 13)) {
		return time.Time{}, fmt.Errorf("invalid syntax: %q", b)
	}
	var n int64
	for _, c := range b {
		if !('0' <= c && c <= '9') {
			return time.Time{}, fmt.Errorf("invalid syntax: %q", b)
		}
		n *= 10
		n += int64(c - byte('0'))
	}
	if len(b) == 13 { // is millis?
		n *= 1e6
	} else {
		n *= 1e9
	}
	return time.Unix(0, n), nil
}

func parseValue(b []byte) (interface{}, error) {
	var v interface{}
	var err error
	if bytes.IndexByte(b, '.') != -1 {
		v, err = parseFloat(b)
	} else {
		v, err = parseInt(b)
	}
	if err != nil {
		err = fmt.Errorf("%v: %q", err, b)
	}
	return v, err
}

func parseInt(b []byte) (int64, error) {
	neg := false
	if len(b) == 0 {
		return 0, fmt.Errorf("invalid syntax")
	}
	if b[0] == '-' {
		neg = true
		b = b[1:]
	}
	var n int64
	for _, c := range b {
		if !('0' <= c && c <= '9') {
			return 0, fmt.Errorf("invalid syntax")
		}
		n *= 10
		n += int64(c - byte('0'))
	}
	if neg {
		n *= -1
	}
	return n, nil
}

func parseFloat(s []byte) (float32, error) {
	n, err := strconv.ParseFloat(string(s), 64)
	return float32(n), err
}

func skipSpace(b []byte) []byte {
	for i, ch := range b {
		if !isSpace(ch) {
			return b[i:]
		}
	}
	return b[len(b):]
}

func skipSpaceNorm(b []byte) []byte {
	if b[0] != ' ' {
		log.Panicf("invalid buf: %s", string(b))
	}
	return b[:1]
}

func skipNonSpace(b []byte) (word, left []byte) {
	for i, ch := range b {
		if isSpace(ch) {
			return b[:i], b[i:]
		}
	}
	return b[:], b[:0]
}

func skipNonSpaceNorm(b []byte) (word, left []byte) {
	i := bytes.IndexByte(b, ' ')
	if i > 0 {
		return b[:i], b[i:]
	}
	return b[:], b[:0]
}

func isSpace(ch byte) bool { return ch == ' ' || ch == '\t' }
