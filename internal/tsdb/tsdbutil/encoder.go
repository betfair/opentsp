// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdbutil

import (
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"opentsp.org/internal/tsdb"
)

// Encoder encodes a data point. If an error is encountered, it is handled
// in implementation-specific way.
type Encoder interface {
	Encode(*tsdb.Point)
}

type execEncoder struct {
	cmd *exec.Cmd
	enc *tsdb.Encoder
}

func NewExecEncoder() Encoder {
	e := new(execEncoder)
	e.reset()
	return e
}

func (e *execEncoder) Encode(p *tsdb.Point) {
	for {
		if err := e.enc.Encode(p); err != nil {
			log.Printf("tsdb encode error: %v", err)
			time.Sleep(5 * time.Second)
			e.reset()
			continue
		}
		break
	}
}

func (e *execEncoder) reset() {
	if e.cmd != nil {
		if err := e.cmd.Process.Kill(); err != nil {
			log.Printf("tsdb-submit kill error: %v", err)
		}
		err := e.cmd.Wait()
		log.Printf("tsdb-submit exit cause: %v", err)
		e.cmd.Stdin.(io.Closer).Close()
		e.cmd = nil
	}
	e.enc = nil

	for e.enc == nil {
		cmd := exec.Command("tsdb-submit")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal(err)
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Printf("tsdb-submit exec error: %v", err)
			stdin.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		e.cmd = cmd
		e.enc = tsdb.NewEncoder(stdin)
	}
}

type stdoutEncoder struct {
	enc *tsdb.Encoder
}

func NewStdoutEncoder() Encoder {
	return &stdoutEncoder{tsdb.NewEncoder(os.Stdout)}
}

func (e *stdoutEncoder) Encode(p *tsdb.Point) {
	if err := e.enc.Encode(p); err != nil {
		log.Fatal(err)
	}
}
