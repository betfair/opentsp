// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"bytes"
	"expvar"
	"fmt"
	"time"
)

var (
	statDecoderBytes   = expvar.NewInt("statse.decoder.Bytes")
	statDecoderErrors  = expvar.NewMap("statse.decoder.Errors")
	statDecoderRecords = expvar.NewInt("statse.decoder.Records")
)

// Reader represents source of undecoded header, body pairs.
type Reader interface {
	ReadParts() [2][]byte
}

type Decoder struct {
	r Reader
}

func NewDecoder(r Reader) *Decoder {
	return &Decoder{r}
}

func (dec *Decoder) Decode(ev *Event) error {
	for {
		parts := dec.r.ReadParts()
		if err := parse(ev, parts); err != nil {
			if Debug != nil {
				Debug.Print("parse error: ", err)
			}
			statDecoderErrors.Add("type=Parse", 1)
			continue
		}
		statDecoderRecords.Add(1)
		statDecoderBytes.Add(sum(parts[:]...))
		return nil
	}
}

func sum(parts ...[]byte) int64 {
	var n int
	for _, p := range parts {
		n += len(p)
	}
	return int64(n)
}

func parse(ev *Event, parts [2][]byte) error {
	if len(parts) != 2 {
		return fmt.Errorf("invalid part count=%d, want 2", len(parts))
	}
	return ev.parse(parts[0], parts[1])
}

func (ev *Event) parse(header, body []byte) error {
	version, timestamp, err := parseHeader(header)
	if err != nil {
		return err
	}
	ev.Time = timestamp
	ev.Error = false
	ev.Statistics = ev.statisticsBuf[:0]
	switch version {
	default:
		return fmt.Errorf("unsupported version: %v", version)
	case 2:
		return ev.parseV2(body)
	case 1:
		return ev.parseV1(body)
	}
}

func parseHeader(s []byte) (version int, timestamp time.Time, err error) {
	i := bytes.IndexByte(s, ' ')
	if i == -1 {
		err = fmt.Errorf("missing space separating version and timestamp")
		return
	}
	version, err = parseVersion(s[:i])
	if err != nil {
		return
	}
	timestamp, err = parseTimestamp(s[i+1:])
	if err != nil {
		return
	}
	return
}

func parseVersion(s []byte) (version int, err error) {
	switch {
	default:
		err = fmt.Errorf("invalid version: want at most 2")
	case bytes.Equal(s, []byte{'1'}):
		version = 1
	case bytes.Equal(s, []byte{'2'}):
		version = 2
	}
	return
}

func parseTimestamp(s []byte) (timestamp time.Time, err error) {
	var millis int64
	for _, c := range s {
		if !('0' <= c && c <= '9') {
			err = fmt.Errorf("invalid timestamp: invalid digit %c in %q", c, string(s))
			return
		}
		millis *= 10
		millis += int64(c - '0')
	}
	sec := millis / 1e3
	nsec := (millis % 1e3) * 1e6
	timestamp = time.Unix(sec, nsec)
	return
}

func (ev *Event) parseStatistics(s []byte, isVersion1 bool) {
	for len(s) > 0 {
		var v []byte
		if i := bytes.IndexByte(s, ' '); i >= 0 {
			v = s[:i]
			s = s[i+1:]
		} else {
			v = s
			s = s[:0]
		}
		switch {
		default:
			if isVersion1 {
				// Some V1 senders include tag-like values in
				// the stats section. Prevent confusing error
				// spam by ignoring these unparsable stats-like
				// tags.
				continue
			}
			if Debug != nil {
				Debug.Printf("drop unsupported statistic, v=%q", v)
			}
		case bytes.HasPrefix(v, []byte("err=")):
			v = v[len("err="):]
			ev.Error = bytes.Equal(v, []byte("true"))
		case bytes.HasPrefix(v, []byte("time=")):
			v = v[len("time="):]
			ev.parseStatistic(Time, v)
		case bytes.HasPrefix(v, []byte("ttfb=")):
			v = v[len("ttfb="):]
			ev.parseStatistic(TTFB, v)
		case bytes.HasPrefix(v, []byte("size=")):
			v = v[len("size="):]
			ev.parseStatistic(Size, v)
		}
	}
}

func (ev *Event) parseStatistic(key Key, s []byte) {
	if defined(ev.Statistics, key) {
		if Debug != nil {
			Debug.Printf("drop redefined statistic, key=%v", key)
		}
		return
	}
	value, err := parseFloat(s)
	if err != nil {
		if Debug != nil {
			Debug.Printf("drop unparsable statistic: %v, key=%v", err, key)
		}
		return
	}
	ev.Statistics = append(ev.Statistics, Statistic{
		Key:   key,
		Value: value,
	})
}

func defined(in []Statistic, key Key) bool {
	for _, stat := range in {
		if stat.Key == key {
			return true
		}
	}
	return false
}

// parseFloat is a fast floating point parser that complies with Statse
// protocol spec.
//
// Compared to strconv.ParseFloat, it never allocates, supports fewer syntaxes,
// and operates on []byte instead of string, which saves a conversion.
func parseFloat(s []byte) (float32, error) {
	if len(s) == 0 {
		err := fmt.Errorf("invalid float: empty string")
		return 0, err
	}
	var a int64
	var b int64
	var denominator []byte
	var d = int64(1)
	for i, c := range s {
		if c == '.' {
			denominator = s[i+1:]
			break
		}
		if !('0' <= c && c <= '9') {
			err := fmt.Errorf("invalid float: invalid digit %c in %q", c, string(s))
			return 0, err
		}
		a *= 10
		a += int64(c - '0')
	}
	for _, c := range denominator {
		if !('0' <= c && c <= '9') {
			err := fmt.Errorf("invalid float: invalid digit %c in %q", c, string(s))
			return 0, err
		}
		b *= 10
		b += int64(c - '0')
		d *= 10
	}
	final := float64(a) + (float64(b) / float64(d))
	return float32(final), nil
}
