package relay

import (
	"log"

	"opentsp.org/internal/tsdb"
)

// Pool represents a pool of relay connections.
type Pool struct {
	relays []*Relay
	series tsdb.Series
}

// NewPool creates a pool of relays connections.
func NewPool(configs map[string]*Config, series tsdb.Series) *Pool {
	var relays []*Relay
	for name, cfg := range configs {
		relay, err := NewRelay(name, cfg)
		if err != nil {
			log.Panicf("internal error: %v", err)
		}
		relays = append(relays, relay)
	}
	pool := &Pool{relays, series}
	return pool
}

// Broadcast broadcasts received data points to all relays.
func (p *Pool) Broadcast() {
	for {
		point := p.series.Next()
		for _, relay := range p.relays {
			relay.Submit(point)
		}
		point.Free()
	}
}
