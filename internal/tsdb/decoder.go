// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"bufio"
	"expvar"
	"fmt"
	"io"
	"time"
)

const (
	decoderMaxSeries     = 1000000
	decoderMaxAge        = 15 * time.Minute
	decoderMaxStep       = 24 * time.Hour
	decoderCleanupEveryN = 100000
)

var (
	statDecoderBytes  = expvar.NewInt("tsdb.decoder.Bytes")
	statDecoderNanos  = expvar.NewInt("tsdb.decoder.Nanos")
	statDecoderErrors = expvar.NewMap("tsdb.decoder.Errors")
)

type Decoder struct {
	r                *bufio.Reader
	bySeries         map[string]*streamState
	cleanupCountdown int
	scratch          [maxLineLength + 1]byte
}

type streamState struct {
	Time int64
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:                bufio.NewReader(r),
		bySeries:         make(map[string]*streamState),
		cleanupCountdown: decoderCleanupEveryN,
	}
}

// Decode decodes the next data point from the input. SyntaxError will be
// returned if an invalid point is encountered. Decoder is valid for use
// despite syntax errors: it can resynchronise.
func (d *Decoder) Decode() (*Point, error) {
	p := getPoint()
	start := time.Now()
	defer func() {
		statDecoderNanos.Add(int64(time.Since(start)))
	}()
	buf, err := d.scan()
	if err != nil {
		statDecoderErrors.Add("type=Read", 1)
		return nil, err
	}
	if err := p.unmarshalText(buf); err != nil {
		statDecoderErrors.Add("type=Syntax", 1)
		return nil, &SyntaxError{err}
	}
	if err := d.validOrder(p); err != nil {
		statDecoderErrors.Add("type=Order", 1)
		return nil, &SyntaxError{err}
	}
	return p, nil
}

func (d *Decoder) scan() ([]byte, error) {
	buf, err := d.r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(buf) > 0 {
		buf = buf[:len(buf)-1]
	}
	statDecoderBytes.Add(int64(len(buf)) + 1)
	return buf, nil
}

// validOrder ensures that the sequence of decoded points is totally ordered in
// time.
func (d *Decoder) validOrder(p *Point) error {
	d.cleanup()
	series := p.appendSeries(d.scratch[:0])
	state, ok := d.bySeries[string(series)]
	if !ok {
		if len(d.bySeries) == decoderMaxSeries {
			return fmt.Errorf("too many time series (>%d)", decoderMaxSeries)
		}
		d.bySeries[string(series)] = &streamState{p.time}
		return nil
	}
	t0 := time.Unix(0, state.Time)
	t1 := time.Unix(0, p.time)
	switch step := t1.Sub(t0); {
	default:
		// ok
	case step < 0:
		return fmt.Errorf("order error: got time %d, want at least %d, in series %q",
			t1.Unix(), t0.Add(maxTimePrecision).Unix(), series)
	case step == 0:
		return fmt.Errorf("order error: collision at time %d, in series %q",
			t0.Unix(), series)
	case step > decoderMaxStep:
		return fmt.Errorf("order error: stepped too far into the future (%s>%s), in series %q",
			step, decoderMaxStep, series)
	}
	state.Time = p.time
	return nil
}

func (d *Decoder) cleanup() {
	d.cleanupCountdown--
	if d.cleanupCountdown > 0 {
		return
	}
	d.cleanupCountdown = decoderCleanupEveryN
	deadline := time.Now().Add(-decoderMaxAge).UnixNano()
	for s, state := range d.bySeries {
		if !(state.Time > deadline) {
			delete(d.bySeries, s)
		}
	}
}

// A SyntaxError signals a syntax error in the TSDB input stream.
type SyntaxError struct {
	error
}
