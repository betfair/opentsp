// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package aggregator

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"opentsp.org/cmd/collect-statse/statse"
)

var statStoreEntryCount = expvar.NewInt("aggregator.store.EntryCount")

// MaxBufferLen limits the number of floats per buffer.
const MaxBufferLen = 10000

type store struct {
	m  map[key]*entry
	mu sync.Mutex
}

func newStore() *store {
	return &store{
		m: make(map[key]*entry),
	}
}

func (s *store) Write(events ...*statse.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, event := range events {
		host, err := parseHostTag(event)
		if err != nil {
			if Debug != nil {
				Debug.Print("drop invalid ", event)
			}
			continue
		}
		key := key{
			Metric: event.Metric,
			Tags:   event.Tags,
			Host:   host,
		}
		entry := s.m[key]
		if entry == nil {
			entry = newEntry()
			s.m[key] = entry
			if Debug != nil {
				Debug.Printf("new entry, key=%v", key)
			}
			statStoreEntryCount.Add(1)
		}
		entry.ModTime = now
		if event.Error {
			entry.CountError++
			continue
		}
		entry.CountOkay++
		for _, stat := range event.Statistics {
			entry.Buffer[stat.Key].Append(stat.Value)
		}
	}
}

func (s *store) del(key key) {
	// NB: s.mu must already be held.
	delete(s.m, key)
}

func (s *store) Do(fn func(key, *entry)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.m {
		fn(key, entry)
	}
}

// key is a lookup key of a store entry.
type key struct {
	Metric, Tags, Host string
}

func (key key) String() string {
	tags := key.Tags + " host=" + key.Host
	return fmt.Sprintf("%q", key.Metric+" "+tags)
}

// entry is a store element accumulating statse statistics.
type entry struct {
	entryMetadata
	CountOkay  uint64
	CountError uint64
	Buffer     [statse.MaxKeys]buffer
}

func newEntry() *entry {
	return &entry{}
}

func (e *entry) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "{")
	fmt.Fprintf(buf, "CountOkay:%d CountError:%d", e.CountOkay, e.CountError)
	for key, values := range e.Buffer {
		if values == nil {
			continue
		}
		fmt.Fprintf(buf, " %v:%v", statse.Key(key), values)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

type cleanupJob struct {
	Store   *store
	Timeout time.Duration
}

func (job *cleanupJob) do() {
	now := time.Now()
	job.Store.Do(func(key key, entry *entry) {
		if now.Sub(entry.ModTime) > job.Timeout {
			job.Store.del(key)
			if Debug != nil {
				Debug.Printf("del entry, key=%v", key)
			}
			statStoreEntryCount.Add(-1)
		}
	})
}

// buffer buffers recently seen values.
type buffer []float32

func (buf *buffer) Append(value float32) {
	if len(*buf) < MaxBufferLen {
		*buf = append(*buf, value)
	} else {
		(*buf)[rand.Intn(MaxBufferLen)] = value
	}
}

func (buf *buffer) Reset() {
	*buf = (*buf)[:0]
}

type entryMetadata struct {
	ModTime time.Time
}

// parseHostTag parses and strips the host tag from the provided event.
func parseHostTag(event *statse.Event) (string, error) {
	var tagsArray [128]byte
	tags := append(tagsArray[:0], event.Tags...)
	i := bytes.Index(tags, []byte("host="))
	if i == -1 || (i > 0 && tags[i-1] != ' ') {
		return "", errors.New("missing host tag")
	}
	tag := tags[i:]
	if i := bytes.IndexByte(tag, ' '); i > 0 {
		tag = tag[:i]
	}
	host := string(tag[bytes.IndexByte(tag, '=')+1:])
	if host == "" {
		return "", errors.New("host tag has empty value")
	}
	tagLen := len("host=") + len(host)
	if i == 0 {
		// host is the first tag.
		copy(tags[i:], tags[i+tagLen:])
		tags = tags[:len(tags)-tagLen]
	} else {
		// host is not the first tag, strip the leading space too.
		copy(tags[i-1:], tags[i+tagLen:])
		tags = tags[:len(tags)-tagLen-1]
	}
	event.Tags = string(tags)
	return host, nil
}
