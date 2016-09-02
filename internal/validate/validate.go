// Package validate implements validation routines for untrusted data.
package validate

import (
	"fmt"
	"net"

	"opentsp.org/internal/relay"
	"opentsp.org/internal/tsdb/filter"
)

const (
	maxRules  = 64
	maxRelays = 8
)

func Filter(rules []filter.Rule) ([]filter.Rule, error) {
	if rules == nil {
		block := true
		rules = []filter.Rule{{Block: &block}}
		return rules, nil
	}
	if n := len(rules); n > maxRules {
		err := fmt.Errorf("too many filter rules defined (%d>%d)", n, maxRules)
		return nil, err
	}
	_, err := filter.New(rules...)
	if err != nil {
		err := fmt.Errorf("error creating filter: %v", err)
		return nil, err
	}
	return rules, nil
}

func Relay(configs map[string]*relay.Config) error {
	if configs == nil {
		err := fmt.Errorf("missing setting: Relay")
		return err
	}
	if n := len(configs); n > maxRelays {
		err := fmt.Errorf("too many relays defined: %d > %d", n, maxRelays)
		return err
	}
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ListenAddr(addr string) error {
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return nil
	}
	if _, _, err := net.SplitHostPort(addr + ":4242"); err == nil {
		return nil
	}
	return fmt.Errorf("invalid listen address: %.100q", addr)
}
