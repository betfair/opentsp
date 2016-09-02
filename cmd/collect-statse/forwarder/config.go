// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package forwarder

import (
	"fmt"
	"log"
	"net"

	"opentsp.org/internal/tsdb/filter"
)

var Debug *log.Logger

type Config struct {
	AggregatorHost string
	Filter         []filter.Rule
}

func (c *Config) Validate() error {
	if err := c.validateAggregatorHost(); err != nil {
		return err
	}
	if err := c.validateFilter(); err != nil {
		return err
	}
	return nil
}

func (c *Config) validateAggregatorHost() error {
	if c.AggregatorHost == "" {
		c.AggregatorHost = "localhost"
	}
	if _, err := net.LookupHost(c.AggregatorHost); err != nil {
		return fmt.Errorf("forwarder: %v, AggregatorHost=%.100q", err, c.AggregatorHost)
	}
	return nil
}

func (c *Config) validateFilter() error {
	if c.Filter == nil {
		block := true;
		c.Filter = []filter.Rule{{Block: &block}}
	}
	_, err := filter.New(c.Filter...)
	if err != nil {
		return fmt.Errorf("forwarder: %v", err)
	}
	return nil
}
