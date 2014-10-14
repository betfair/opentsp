// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"fmt"
	"io"
	"time"

	"opentsp.org/internal/tsdb"
)

// decoder is like tsdb.Decoder except it declares idle timeout after a period of
// inactivity.
type decoder struct {
	dec          *tsdb.Decoder
	timeout      time.Duration
	timeoutTimer *time.Timer
	pointChan    chan *tsdb.Point
	errChan      chan error
}

func newDecoder(r io.Reader, timeout time.Duration) *decoder {
	d := &decoder{
		dec:          tsdb.NewDecoder(r),
		timeout:      timeout,
		timeoutTimer: time.NewTimer(timeout),
		pointChan:    make(chan *tsdb.Point),
		errChan:      make(chan error),
	}
	go d.mainloop()
	return d
}

// Decode is like tsdb.Decoder's Decode except it may return *decoderTimeout,
func (d *decoder) Decode() (*tsdb.Point, error) {
	d.timeoutTimer.Reset(d.timeout)
	select {
	case point := <-d.pointChan:
		return point, nil
	case err := <-d.errChan:
		return nil, err
	case <-d.timeoutTimer.C:
		go d.drain()
		err := fmt.Errorf("idle timeout: inactive for %ds", d.timeout.Nanoseconds()/1e9)
		return nil, &decoderTimeout{err}
	}
}

func (d *decoder) mainloop() {
	defer close(d.pointChan)
	for {
		point, err := d.dec.Decode()
		if err != nil {
			switch err.(type) {
			default:
				d.errChan <- err
				return
			case *tsdb.SyntaxError:
				d.errChan <- err
				continue
			}
		}
		d.pointChan <- point
	}
}

// drain discards all data until mainloop returns.
func (d *decoder) drain() {
	for {
		select {
		case _, ok := <-d.pointChan:
			if !ok {
				return
			}
			// discard all points
		case <-d.errChan:
			// discard all errors
		}
	}
}

type decoderTimeout struct {
	error
}
