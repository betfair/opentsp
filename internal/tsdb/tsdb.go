// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package tsdb implements OpenTSDB protocol.
package tsdb

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	maxTimePrecision = 1 * time.Second
	maxTagsPerPoint  = 8
)

type Point struct {
	time       int64
	valueInt   int64
	valueFloat float32
	isFloat    bool
	metric     []byte
	tags       []byte
	// buffers for allocation combining
	metricBuf [128]byte
	tagsBuf   [128]byte
}

func (p *Point) Copy() *Point {
	point := *p
	point.metric = append(point.metricBuf[:0], p.metric...)
	point.tags = append(point.tagsBuf[:0], p.tags...)
	return &point
}

// NewPoint accepts any numeric or string value.
func NewPoint(time time.Time, value interface{}, metric string, keyval ...string) (*Point, error) {
	p := getPoint()
	if err := p.setTime(time); err != nil {
		return nil, fmt.Errorf("tsdb: invalid time: %v", err)
	}
	if err := p.SetValue(value); err != nil {
		return nil, err
	}
	if err := p.SetMetric([]byte(metric)); err != nil {
		return nil, err
	}
	buf := make([]byte, 0, 128)
	for i, s := range keyval {
		if i%2 == 0 {
			buf = append(buf, ' ')
			buf = append(buf, s...)
		} else {
			buf = append(buf, '=')
			buf = append(buf, s...)
		}
	}
	if err := p.setTags(buf, skipNonSpaceNorm); err != nil {
		return nil, fmt.Errorf("tsdb: invalid tags: %v", err)
	}
	return p, nil
}

func (p *Point) reset() {
	p.metric = p.metricBuf[:0]
	p.tags = p.tagsBuf[:0]
	p.isFloat = false
}

func (p *Point) Equal(q *Point) bool {
	return p.time == q.time &&
		p.Value() == q.Value() &&
		bytes.Equal(p.metric, q.metric) &&
		bytes.Equal(p.tags, q.tags)
}

func (p *Point) String() string {
	tags := "nil"
	if len(p.tags) > 0 {
		tags = fmt.Sprintf("%q", string(p.tags[1:]))
	}
	t := time.Unix(0, p.time)
	return fmt.Sprintf("{Time:%d Value:%T(%v) Metric:%q Tags:%s}",
		t.Unix(), p.Value(), p.Value(), string(p.metric), tags)
}

func (p *Point) Time() time.Time {
	return time.Unix(0, p.time)
}

func (p *Point) setTime(time time.Time) error {
	t, err := validateTime(time)
	if err != nil {
		return err
	}
	p.time = t
	return nil
}

func (p *Point) Value() interface{} {
	if p.isFloat {
		return p.valueFloat
	}
	return p.valueInt
}

func (p *Point) SetValue(value interface{}) error {
	var err error
	switch v := value.(type) {
	default:
		err = p.setValue(v)
	case []byte:
		err = p.setValueBytes(v)
	case string:
		err = p.setValueBytes([]byte(v))
	}
	if err != nil {
		err = fmt.Errorf("tsdb: invalid value: %v", err)
	}
	return err
}

func (p *Point) setValueBytes(b []byte) error {
	if b == nil {
		return fmt.Errorf("nil")
	}
	v, err := parseValue(b)
	if err != nil {
		return err
	}
	return p.setValue(v)
}

func (p *Point) setValue(value interface{}) error {
	v, err := validateValue(value)
	if err != nil {
		return err
	}
	p.setValueValid(v)
	return nil
}

func (p *Point) setValueValid(v interface{}) {
	n, ok := v.(float32)
	if ok {
		p.isFloat = true
		p.valueFloat = n
		return
	}
	p.isFloat = false
	p.valueInt = v.(int64)
}

func (p *Point) Metric() []byte {
	return p.metric
}

func (p *Point) SetMetric(s []byte) error {
	err := p.setMetric(s)
	if err != nil {
		return fmt.Errorf("tsdb: invalid metric: %v", err)
	}
	return err
}

func (p *Point) setMetric(b []byte) error {
	if err := validMetric(b); err != nil {
		return err
	}
	p.metric = append(p.metric[:0], b...)
	return nil
}

var space = []byte{' '}

// Tag returns the value associated with the given tag key.
// It returns an nil slice if tag was not found.
func (p *Point) Tag(key []byte) []byte {
	if bytes.Contains(key, space) {
		return nil
	}
	tags := p.tags
	for {
		i := bytes.Index(tags, key)
		if i == -1 {
			break
		}
		// Does any data follow the key? We expect at least the
		// '=' delimiter.
		if len(tags[i+len(key):]) == 0 {
			break
		}
		// Did we find the tag key exactly? Check both delimiters.
		if tags[i-1] == ' ' && tags[i+len(key)] == '=' {
			v := tags[i+len(key)+1:]
			if i := bytes.IndexByte(v, ' '); i > 0 {
				v = v[:i]
			}
			return v
		}
		// Skip to the next tag.
		i = bytes.IndexByte(tags[1:], ' ')
		if i == -1 {
			break
		}
		tags = tags[i+1:]
	}
	return nil
}

func (p *Point) Tags() []string {
	if len(p.tags) == 0 {
		return nil
	}
	tags := strings.Replace(string(p.tags[1:]), "=", " ", -1)
	return strings.Fields(tags)
}

func (p *Point) SetTags(keyval ...[]byte) error {
	if len(keyval) > 0 && len(keyval)%2 == 1 {
		return fmt.Errorf("tsdb: SetTags: got tag without value")
	}
	// Temp area.
	tags := make([]byte, 0, 128)
	keys := make([][]byte, 0, maxTagsPerPoint)
	seen := make([][]byte, 0, maxTagsPerPoint)
	// Write the new tags.
	for i := 0; i < len(keyval); i += 2 {
		k := keyval[i]
		v := keyval[i+1]
		if err := validTag(k, v); err != nil {
			return fmt.Errorf("tsdb: SetTags: invalid tag: %v", err)
		}
		tags = append(tags, ' ')
		tags = append(tags, k...)
		tags = append(tags, '=')
		tags = append(tags, v...)
		// Keep track of the new tags.
		keys = append(keys, k)
		// Keep track of all tags.
		seen = append(seen, k)
	}
	// Append old tags.
	buf := p.tags
Next:
	for len(buf) > 0 {
		var k, v []byte
		k, v, buf = nextTag(buf)
		for _, setKey := range keys {
			if bytes.Equal(k, setKey) {
				continue Next
			}
		}
		tags = append(tags, ' ')
		tags = append(tags, k...)
		tags = append(tags, '=')
		tags = append(tags, v...)
		// Keep track of all tags.
		seen = append(seen, k)
	}
	if ntags := len(seen); ntags > maxTagsPerPoint {
		return fmt.Errorf("tsdb: SetTags: too many tags (%d>%d)", ntags, maxTagsPerPoint)
	}
	p.tags = append(p.tags[:0], tags...)
	return nil
}

func nextTag(buf []byte) (k, v, rest []byte) {
	i := bytes.IndexByte(buf, '=')
	j := bytes.IndexByte(buf[1:], ' ')
	if j == -1 {
		k = buf[1:i]
		v = buf[i+1:]
		rest = nil
	} else {
		k = buf[1:i]
		v = buf[i+1 : j+1]
		rest = buf[j+1:]
	}
	return
}

func (p *Point) XAppendTags(nameval ...string) error {
	for i, s := range nameval {
		if i%2 == 0 {
			p.tags = append(p.tags, ' ')
			p.tags = append(p.tags, s...)
		} else {
			p.tags = append(p.tags, '=')
			p.tags = append(p.tags, s...)
		}
	}
	return nil
}

func (p *Point) setTags(buf []byte, skipNSP func([]byte) ([]byte, []byte)) error {
	seen := make([][]byte, 0, maxTagsPerPoint)
	tags := p.tags[:0]
	for {
		buf = skipSpace(buf)
		if len(buf) == 0 {
			break
		}
		var tag []byte
		tag, buf = skipNSP(buf)
		i := bytes.IndexByte(tag, '=')
		if i == -1 {
			return fmt.Errorf("truncated tags list")
		}
		k := tag[:i]
		v := tag[i+1:]
		if err := validTag(k, v); err != nil {
			return err
		}
		for _, seenk := range seen {
			if bytes.Equal(k, seenk) {
				return fmt.Errorf("duplicate tag key %q", k)
			}
		}
		seen = append(seen, k)
		tags = append(tags, ' ')
		tags = append(tags, k...)
		tags = append(tags, '=')
		tags = append(tags, v...)
	}
	if ntags := len(seen); ntags > maxTagsPerPoint {
		return fmt.Errorf("too many tags (%d>%d)", ntags, maxTagsPerPoint)
	}
	if len(seen) > 0 {
		p.tags = tags
	} else {
		p.tags = nil
	}
	return nil
}

func validateTime(time time.Time) (int64, error) {
	if time.IsZero() {
		return -1, fmt.Errorf("zero struct")
	}
	t := time.Truncate(maxTimePrecision)
	return t.UnixNano(), nil
}

func validateValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, fmt.Errorf("nil")
	}
	switch n := v.(type) {
	default:
		return nil, fmt.Errorf("type %T", n)
	case int:
		v = int64(n)
	case int8:
		v = int64(n)
	case int16:
		v = int64(n)
	case int32:
		v = int64(n)
	case int64:
		// ok
	case uint:
		v = int64(n)
	case uint8:
		v = int64(n)
	case uint16:
		v = int64(n)
	case uint32:
		v = int64(n)
	case uint64:
		v = int64(n)
	case float32:
		// ok
	case float64:
		v = float32(n)
	}
	if n, ok := v.(float32); ok {
		if math.IsInf(float64(n), 0) {
			return nil, fmt.Errorf("Inf")
		}
		if math.IsNaN(float64(n)) {
			return nil, fmt.Errorf("NaN")
		}
	}
	return v, nil
}

func validMetric(b []byte) error {
	if err := validText(b); err != nil {
		return err
	}
	return nil
}

func validTag(k, v []byte) error {
	if err := validText(k); err != nil {
		return fmt.Errorf("%v, in tag %q", err, string(k)+"="+string(v))
	}
	if err := validText(v); err != nil {
		return fmt.Errorf("%v, in tag %q", err, string(k)+"="+string(v))
	}
	return nil
}

func validText(b []byte) error {
	if len(b) == 0 {
		return fmt.Errorf("empty string")
	}
	for len(b) > 0 {
		r, sz := utf8.DecodeRune(b)
		b = b[sz:]
		switch {
		default:
			return fmt.Errorf("invalid character %q", string(r))
		case 'a' <= r && r <= 'z':
			// ok
		case 'A' <= r && r <= 'Z':
			// ok
		case unicode.IsDigit(r), unicode.IsLetter(r):
			// ok
		case r == '-', r == '_', r == '.', r == '/':
			// ok
		}
	}
	return nil
}

func newPoint() interface{} { return new(Point) }

var pool = sync.Pool{New: newPoint}

// getPoint allocates a new point from the internal sync.Pool.
func getPoint() *Point {
	point := pool.Get().(*Point)
	point.reset()
	return point
}

// Free returns Point's memory into an internal free list.  Free is safe to call
// after a final Encode or Put.
func (p *Point) Free() {
	pool.Put(p)
}
