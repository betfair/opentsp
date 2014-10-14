// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"fmt"
	"log"
	"time"

	"opentsp.org/internal/tsdb"
)

var Debug *log.Logger

var tsdbChan chan<- *tsdb.Point

// Init initialises the package. Must be run at init time.
func Init(tsdbChan_ chan<- *tsdb.Point) {
	tsdbChan = tsdbChan_
}

type Config struct {
	SnapshotInterval string
}

func (c *Config) Validate() error {
	if err := c.validateSnapshotInterval(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateSnapshotInterval() error {
	if c.SnapshotInterval == "" {
		c.SnapshotInterval = "10s"
	}
	interval, err := time.ParseDuration(c.SnapshotInterval)
	if err != nil {
		return fmt.Errorf("aggregator: %v, SnapshotInterval=%.100q", err, c.SnapshotInterval)
	}
	if interval < 1*time.Second {
		return fmt.Errorf("aggregator: interval too short: want at least 1s, got %.100q: ",
			c.SnapshotInterval)
	}
	return nil
}
