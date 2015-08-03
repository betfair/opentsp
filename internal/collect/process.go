// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package collect implements fan-in from local plugin programs.
package collect

import (
	"bufio"
	"expvar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"opentsp.org/internal/tsdb"
)

var Debug *log.Logger

const (
	retryDelay      = 5 * time.Second  // between retries in case of exec errors
	repairDelay     = 5 * time.Second  // between repairs of a crashed plugin
	rescheduleDelay = 1 * time.Hour    // between reschedules of a plugin
	idleTimeout     = 10 * time.Minute // if breached, the process is killed
	exitTimeout     = 1 * time.Second  // if breached, a warning is printed
)

var (
	statErrors       = expvar.NewMap("collect.Errors")
	statPoints       = expvar.NewInt("collect.Points")
	statProcessCount = expvar.NewInt("collect.ProcessCount")
	statQueue        = expvar.NewMap("collect.Queue")
)

// process represents a running collection program.
type process struct {
	path       string
	cmd        *exec.Cmd
	closePipes func()
	killChan   chan bool
	Start      time.Time
	Exit       chan error
}

// startProcess starts a new process corresponding to the given directory path.
func startProcess(path string, w chan<- *tsdb.Point) *process {
	p := &process{
		path:     path,
		killChan: make(chan bool, 1),
		Start:    time.Now(),
		Exit:     make(chan error, 1),
	}
	stdoutRead, stdoutWrite, err := os.Pipe()
	if err != nil {
		statErrors.Add("type=Start", 1)
		p.Exit <- &startError{err}
		return p
	}
	stderrRead, stderrWrite, err := os.Pipe()
	if err != nil {
		stdoutRead.Close()
		stdoutWrite.Close()
		statErrors.Add("type=Start", 1)
		p.Exit <- &startError{err}
		return p
	}
	p.cmd = exec.Command(path)
	p.cmd.Env = safeEnviron()
	p.cmd.Stdout = stdoutWrite
	p.cmd.Stderr = stderrWrite
	if err := p.cmd.Start(); err != nil {
		stdoutWrite.Close()
		stderrWrite.Close()
		stdoutRead.Close()
		stderrRead.Close()
		statErrors.Add("type=Start", 1)
		p.Exit <- &startError{err}
	} else {
		statProcessCount.Add(1)
		stdoutWrite.Close()
		stderrWrite.Close()
		p.closePipes = func() {
			stdoutRead.Close()
			stderrRead.Close()
		}
		go p.decode(stdoutRead, w)
		go p.stderrLogger(stderrRead)
		go p.handleKill()
	}
	return p
}

func (p *process) Kill() {
	select {
	case p.killChan <- true:
		// ok
	default:
		// already marked for killing
	}
}

func (p *process) Printf(format string, arg ...interface{}) {
	dir, file := filepath.Split(p.path)
	pid := filepath.Join(filepath.Base(dir), file)
	format = fmt.Sprintf("%s: %s", pid, format)
	log.Printf(format, arg...)
}

// decode decodes data points errors available via stdout.
func (p *process) decode(r io.Reader, w chan<- *tsdb.Point) {
	dec := newDecoder(r, idleTimeout)
	for {
		point, err := dec.Decode()
		if err != nil {
			switch err.(type) {
			default:
				return
			case *tsdb.SyntaxError:
				p.Printf("%v", err)
				continue
			case *decoderTimeout:
				p.Printf("kill (%v)", err)
				p.Kill()
				return
			}
		}
		statPoints.Add(1)
		select {
		case w <- point:
			// ok
		default:
			statErrors.Add("type=Enqueue", 1)
			w <- point
		}
	}
}

// stderrLogger logs process errors available via stderr.
func (p *process) stderrLogger(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		statErrors.Add("type=Stderr", 1)
		p.Printf("stderr: %s", scanner.Text())
	}
	_ = scanner.Err()
}

// handleKill handles termination of the process.
func (p *process) handleKill() {
	wait := make(chan error)
	go func() {
		err := p.cmd.Wait()
		if err == nil {
			err = &cleanExit{}
		}
		wait <- err
	}()
	select {
	case err := <-wait:
		// Process died on its own, no need to kill.
		p.closePipes()
		p.Exit <- err
	case <-p.killChan:
		// Process alive but required to die.
		if err := p.cmd.Process.Kill(); err != nil {
			p.Printf("kill error: %v", err)
		}
		// Cause "broken pipe" error on all pipe writers. The set of all writers
		// may be larger than the process just killed (think fd inheritance).
		p.closePipes()
		// Block until the process is reaped.
		var err error
		for {
			select {
			case err = <-wait:
				// ok
			case <-time.After(exitTimeout):
				p.Printf("slow exit, still waiting...")
				continue
			}
			break
		}
		p.Exit <- err
	}
	statProcessCount.Add(-1)
}

type startError struct {
	error
}

type cleanExit struct{}

func (*cleanExit) Error() string {
	return "exit status 0"
}

// restart returns a schedule for process restart based on its exit status.
func restart(process *process, err error) <-chan time.Time {
	delay := retryDelay
	switch err.(type) {
	default:
		if reschedule(err) {
			process.Printf("%v, restart in %ds", err, rescheduleDelay.Nanoseconds()/1e9)
			delay = rescheduleDelay
		} else {
			statErrors.Add("type=Crash", 1)
			process.Printf("%v (after running for %ds)", err, time.Since(process.Start).Nanoseconds()/1e9)
			delay = repairDelay
		}
	case *cleanExit:
		process.Printf("%v (after running for %ds)", err, time.Since(process.Start).Nanoseconds()/1e9)
	case *startError:
		statErrors.Add("type=Start", 1)
		process.Printf("%v", err)
	}
	return time.After(delay)
}

// safeEnviron is like os.Environ except it excludes variables that cause
// confusion when inherited.
func safeEnviron() []string {
	blacklist := map[string]bool{
		"GODEBUG":    true,
		"GOGC":       true,
		"GOMAXPROCS": true,
	}
	var environ []string
	for _, kv := range os.Environ() {
		k := kv[:strings.Index(kv, "=")]
		if blacklist[k] {
			continue
		}
		environ = append(environ, kv)
	}
	return environ
}
