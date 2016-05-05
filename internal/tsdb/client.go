// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdb

import (
	"bytes"
	"hash"
	"hash/fnv"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const clientDialInterval = 100 * time.Millisecond

const (
	clientMaxQueue = 100000
)

type ClientConfig struct {
	DropRepeats     bool
	MaxConnsPerHost int
	// The slice is valid only for the duration of the Drop call.
	Drop func([]byte)
}

// Client represents a connection pool to the TSDB server.
type Client struct {
	once     sync.Once
	hosts    string
	Dial     func(string) (net.Conn, error)
	dialRate *time.Ticker
	cmd      chan cmd
	config   *ClientConfig
	repeat   *repeatTester
	up       []*clientConn
	upMu     sync.RWMutex
	upHash   hash.Hash32
}

// NewClient returns a TSDB client. The client supervises internal connection pool.
// Network errors are handled by re-connecting with exponential backoff.
func NewClient(hosts string, config *ClientConfig) *Client {
	c := &Client{
		hosts:    hosts,
		Dial:     newDial(),
		dialRate: time.NewTicker(clientDialInterval),
		cmd:      make(chan cmd, clientMaxQueue),
		repeat:   newRepeatTester(),
		config:   config,
		upHash:   fnv.New32(),
	}
	go c.mainloop()
	return c
}

// Put writes a data point to the server. It never blocks on I/O.
func (c *Client) Put(point *Point) {
	c.once.Do(func() {
		c.dialAll()
	})
	if c.config.DropRepeats {
		isRepeat, held := c.repeat.Test(point)
		if held != nil {
			c.send(held)
		}
		if isRepeat {
			return
		}
	}
	cmd := point.put()
	c.send(cmd)
}

func (c *Client) dialAll() {
	for _, addr := range parseHosts(c.hosts) {
		for i := 0; i < c.config.MaxConnsPerHost; i++ {
			go c.dial(addr)
		}
	}
}

func (c *Client) send(cmd cmd) {
	select {
	case c.cmd <- cmd:
		// ok
	default:
		c.drop(cmd.Point())
		cmd.Free()
	}
}

func (c *Client) drop(lines []byte) {
	if c.config.Drop == nil {
		return
	}
	c.config.Drop(lines)
}

// Queue returns the client queue. It's exposed as receive-only to enable
// only metrics gathering (length, capacity).
func (c *Client) Queue() <-chan cmd {
	return c.cmd
}

func (c *Client) mainloop() {
	var cmd cmd
	var conn *clientConn
	var err error
	for cmd = range c.cmd {
		conn = c.lookupConn(cmd)
		if err = conn.Put(cmd); err != nil {
			c.error(conn, err)
		}
	}
}

func (c *Client) lookupConn(cmd cmd) *clientConn {
	var up *clientConn
	for {
		c.upMu.RLock()
		if len(c.up) > 0 {
			up = c.up[cmd.SeriesHash(c.upHash)%len(c.up)]
		}
		c.upMu.RUnlock()
		if up != nil {
			return up
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) dial(addr string) {
	for {
		<-c.dialRate.C
		conn, err := c.Dial(addr)
		if err != nil {
			continue
		}
		c.upMu.Lock()
		c.up = append(c.up, newClientConn(conn, addr))
		c.upMu.Unlock()
		break
	}
}

func (c *Client) error(conn *clientConn, _ error) {
	c.upMu.Lock()
	c.up = remove(c.up, conn)
	c.upMu.Unlock()
	conn.Close()
	c.drop(conn.Pending.Bytes())
	go c.dial(conn.Addr)
}

// put returns a put command corresponding to this Point.
func (p *Point) put() cmd {
	cmd := cmdPool.Get().(cmd)
	cmd.Reset()
	cmd = append(cmd, "put "...)
	cmd = p.append(cmd)
	cmd = append(cmd, '\n')
	return cmd
}

var cmdPool = sync.Pool{New: newCmd}

// cmd holds a put command to TSDB server.
type cmd []byte

func newCmd() interface{} {
	return make(cmd, 0, 256)
}

func (c cmd) String() string {
	return string(c)
}

func (c *cmd) Reset() {
	*c = (*c)[:0]
}

// Point returns a marshalled version of put command's point.
func (c cmd) Point() []byte {
	return c[len("put ") : len(c)-1]
}

// SeriesHash returns a hash of the metric, tags pair. The hash falls in uint16
// range.
func (c cmd) SeriesHash(hash hash.Hash32) int {
	hash.Reset()
	buf := c.Point()
	// include Metric
	i := bytes.IndexByte(buf, ' ')
	// this should never happen, as an empty Point representation has
	// 2 space chars, one before the timestamp and another before the value
	// e.g. " 0 0"
	if i < 0 {
		log.Printf("SeriesHash: invalid buf: %s", string(c.Point()))
		return 0 // return a default hash
	}
	hash.Write(buf[:i])
	buf = buf[i+1:]
	// exclude Time
	i = bytes.IndexByte(buf, ' ')
	buf = buf[i+1:]
	// exclude Value
	i = bytes.IndexByte(buf, ' ')
	if i < 0 {
		log.Printf("SeriesHash: invalid buf: %s", string(c.Point()))
		return 0 // return a default hash
	}
	// include Tags
	hash.Write(buf[i:])
	// Sum
	sum := hash.Sum32()
	return int(uint16(sum))
}

func (c cmd) Free() {
	cmdPool.Put(c)
}

func remove(list []*clientConn, conn *clientConn) []*clientConn {
	var tmp []*clientConn
	for _, c := range list {
		if c != conn {
			tmp = append(tmp, c)
		}
	}
	return tmp
}

func parseHosts(s string) []string {
	return strings.Split(s, ",")
}
