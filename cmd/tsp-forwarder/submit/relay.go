// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package submit

import (
	"bytes"
	"expvar"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"opentsp.org/internal/tsdb"
)

const (
	oqfDrop       = "Drop"
	oqfDropAndLog = "DropAndLog"
)

var (
	statRelayCurrEstab = expvar.NewMap("relay.CurrEstab")
	statRelayErrors    = expvar.NewMap("relay.Errors")
	statRelayQueue     = expvar.NewMap("relay.Queue")
)

type RelayConfig struct {
	DropRepeats     bool
	Host            string
	MaxConnsPerHost *int
	OnQueueFull     string
}

func (c *RelayConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("invalid relay: missing Host")
	}
	switch c.OnQueueFull {
	default:
		return fmt.Errorf("invalid OnQueueFull: %q", c.OnQueueFull)
	case "":
		c.OnQueueFull = oqfDrop
	case oqfDrop, oqfDropAndLog:
		// ok
	}
	switch max, defaultMax := c.MaxConnsPerHost, 1; {
	default:
		return fmt.Errorf("MaxConnsPerHost out of range: %d", *max)
	case max == nil:
		c.MaxConnsPerHost = &defaultMax
	case *max < 16:
		// ok
	}
	return nil
}

type Relay struct {
	name   string
	host   string
	drop   func([]byte)
	client *tsdb.Client
}

// NewRelay returns a new relay.
func NewRelay(name string, config *RelayConfig) (*Relay, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("relay %s: %v", name, err)
	}
	r := &Relay{
		name: name,
		host: config.Host,
	}
	r.drop = drop(name)
	if config.OnQueueFull == oqfDropAndLog {
		r.drop = logLost(name, r.drop)
	}
	r.client = tsdb.NewClient(config.Host, &tsdb.ClientConfig{
		DropRepeats:     config.DropRepeats,
		MaxConnsPerHost: *config.MaxConnsPerHost,
		Drop:            r.drop,
	})
	r.client.Dial = dial(name, r.client.Dial)
	queue := r.client.Queue()
	statRelayQueue.Set("relay="+name, expvar.Func(func() interface{} {
		return len(queue)
	}))
	return r, nil
}

// Submit submits the given point to the relay. It does not block in network calls
// to the relay host. Not safe for concurrent use.
func (r *Relay) Submit(point *tsdb.Point) {
	r.client.Put(point)
}

type dialFunc func(string) (net.Conn, error)

func dial(name string, fn dialFunc) dialFunc {
	return func(addr string) (net.Conn, error) {
		conn, err := fn(addr)
		if err != nil {
			statRelayErrors.Add("type=Dial relay="+name, 1)
			log.Printf("relay %s: %v", name, err)
			return nil, err
		}
		statRelayCurrEstab.Add("relay="+name, 1)
		return connMonitor{conn, name}, nil
	}
}

type dropFunc func([]byte)

func drop(name string) dropFunc {
	return func(buf []byte) {
		npoints := bytes.Count(buf, []byte{'\n'})
		statRelayErrors.Add("type=Drop relay="+name, int64(npoints))
	}
}

func logLost(name string, fn dropFunc) dropFunc {
	mu := sync.Mutex{}
	return func(buf []byte) {
		fn(buf)
		mu.Lock()
		for len(buf) > 0 {
			i := bytes.IndexByte(buf, '\n')
			point := buf[:i]
			buf = buf[i+1:]
			fmt.Fprintf(os.Stdout, "relay %s: lost: %s\n", name, point)
		}
		mu.Unlock()
	}
}

// connMonitor logs connection errors for the given relay.
type connMonitor struct {
	net.Conn
	Relay string
}

func (cm connMonitor) Read(buf []byte) (int, error) {
	n, err := cm.Conn.Read(buf)
	if err != nil {
		log.Printf("relay %s: %v", cm.Relay, err)
	}
	return n, err
}

func (cm connMonitor) Write(buf []byte) (int, error) {
	n, err := cm.Conn.Write(buf)
	if err != nil {
		log.Printf("relay %s: %v", cm.Relay, err)
	}
	return n, err
}

func (cm connMonitor) Close() error {
	err := cm.Conn.Close()
	if err != nil {
		log.Printf("relay %s: %v", cm.Relay, err)
	} else {
		statRelayCurrEstab.Add("relay="+cm.Relay, -1)
	}
	return err
}
