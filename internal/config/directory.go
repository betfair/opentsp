// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// directoryWatchInterval is the interval between Directory scans.
const directoryWatchInterval = 1 * time.Second

// directoryEventRate limits per-second rate of delivery of Directory events.
const directoryEventRate = 10

// DirectoryEvent represents addition/removal event of the given file.
type DirectoryEvent struct {
	Target   string
	IsCreate bool
	IsRemove bool
	IsModify bool
}

func (event DirectoryEvent) String() string {
	if event.IsCreate {
		return fmt.Sprintf("Create{%q}", event.Target)
	}
	if event.IsRemove {
		return fmt.Sprintf("Remove{%q}", event.Target)
	}
	if event.IsModify {
		return fmt.Sprintf("Modify{%q}", event.Target)
	}
	return fmt.Sprintf("InvalidEvent{%q}", event.Target)
}

type directoryEntry struct {
	Mode    os.FileMode
	ModTime time.Time
}

// Directory is a watcher of directory changes.
type Directory struct {
	path      string
	last      map[string]directoryEntry
	event     chan DirectoryEvent
	eventRate <-chan time.Time
	C         <-chan DirectoryEvent
	stop      chan bool
}

// WatchDirectory watches the given directory for updates. The scan is
// not recursive.
func WatchDirectory(path string) *Directory {
	ch := make(chan DirectoryEvent)
	dir := &Directory{
		path:      path,
		last:      make(map[string]directoryEntry),
		event:     ch,
		eventRate: time.Tick(1 * time.Second / directoryEventRate),
		C:         ch,
		stop:      make(chan bool),
	}
	go dir.watch()
	return dir
}

func (dir *Directory) Stop() {
	dir.stop <- true
	<-dir.stop
}

func (dir *Directory) watch() {
	defer close(dir.stop)
	ticker := time.NewTicker(directoryWatchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			dir.diff()
		case <-dir.stop:
			return
		}
	}
}

func (dir *Directory) diff() {
	have := dir.walk()
	for path, now := range have {
		last, ok := dir.last[path]
		if !ok {
			dir.eventCreate(path)
			continue
		}
		if now != last {
			dir.eventModify(path)
			continue
		}
	}
	for path := range dir.last {
		if _, ok := have[path]; !ok {
			dir.eventRemove(path)
			continue
		}
	}
	dir.last = have
}

func (dir *Directory) walk() map[string]directoryEntry {
	have := make(map[string]directoryEntry)
	filepath.Walk(dir.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("config: directory walk error: %v", err)
			return nil
		}
		if path == dir.path {
			return nil
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		if isDotfile(path) {
			return nil
		}
		have[path] = directoryEntry{info.Mode(), info.ModTime()}
		return nil
	})
	return have
}

func (dir *Directory) eventCreate(path string) {
	<-dir.eventRate
	dir.event <- DirectoryEvent{
		Target:   path,
		IsCreate: true,
	}
}

func (dir *Directory) eventRemove(path string) {
	<-dir.eventRate
	dir.event <- DirectoryEvent{
		Target:   path,
		IsRemove: true,
	}
}

func (dir *Directory) eventModify(path string) {
	<-dir.eventRate
	dir.event <- DirectoryEvent{
		Target:   path,
		IsModify: true,
	}
}
