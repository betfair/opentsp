// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import "time"

type scheduler struct {
	store        *store
	snapshotTick <-chan time.Time
	cleanupTick  <-chan time.Time
}

func newScheduler(store *store, interval time.Duration) *scheduler {
	s := &scheduler{
		store:        store,
		snapshotTick: time.Tick(interval),
		cleanupTick:  time.Tick(cleanupInterval),
	}
	go s.loop()
	return s
}

const cleanupInterval = 1 * time.Minute

func (s *scheduler) loop() {
	for {
		select {
		case t := <-s.snapshotTick:
			s.snapshot(t)
		case <-s.cleanupTick:
			s.cleanup()
		}
	}
}

// snapshot creates data points based on the current snapshot of the store.
func (s *scheduler) snapshot(t time.Time) {
	job := snapshotJob{
		Store: s.store,
		Time:  t,
	}
	job.do()
	for _, point := range job.Output {
		tsdbChan <- point
	}
}

const defaultTimeout = 1 * time.Minute

// cleanup removes inactive store entries.
func (s *scheduler) cleanup() {
	job := cleanupJob{
		Store:   s.store,
		Timeout: defaultTimeout,
	}
	job.do()
}
