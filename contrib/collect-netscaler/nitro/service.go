// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package nitro

import (
	"fmt"
)

// ConfigService represents the configuration section of the Nitro API.
type ConfigService struct {
	client *Client
}

func (s *ConfigService) Get(entity string) (*ResponseConfig, error) {
	resp := new(ResponseConfig)
	if err := s.client.do(resp, "config/"+entity); err != nil {
		return nil, fmt.Errorf("nitro: Config: %v", err)
	}
	return resp, nil
}

// StatService represents the statistics section of the Nitro API.
type StatService struct {
	client *Client
}

// Get retrieves statistics for the provided entity.
func (s *StatService) Get(entity string) (*ResponseStat, error) {
	resp := new(ResponseStat)
	if err := s.client.do(resp, "stat/"+entity); err != nil {
		return nil, fmt.Errorf("nitro: Stat: %v", err)
	}
	return resp, nil
}
