// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"expvar"
	"io"
	"time"
)

var (
	statEncoderBytes  = expvar.NewInt("tsdb.encoder.Bytes")
	statEncoderNanos  = expvar.NewInt("tsdb.encoder.Nanos")
	statEncoderErrors = expvar.NewMap("tsdb.encoder.Errors")
)

type Encoder struct {
	w       io.Writer
	scratch [maxLineLength + 1]byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(p *Point) error {
	start := time.Now()
	defer func() {
		statEncoderNanos.Add(int64(time.Since(start)))
	}()
	buf := e.scratch[:0]
	buf = p.append(buf)
	buf = append(buf, '\n')
	_, err := e.w.Write(buf)
	if err != nil {
		statEncoderErrors.Add("type=Write", 1)
		return err
	}
	statEncoderBytes.Add(int64(len(buf)))
	return nil
}
