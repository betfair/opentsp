package collect

import (
	"expvar"
	"log"
	"time"

	"opentsp.org/internal/config"
	"opentsp.org/internal/tsdb"
)

const (
	// MaxQueue limits the number of points queued in a Pool.
	MaxQueue = 10000

	// MaxProc limits the number of processes running in parallel.
	MaxProc = 128
)

// Pool represents a pool of plugin processes.
type Pool struct {
	C         <-chan *tsdb.Point
	directory *config.Directory
	byPath    map[string]*directoryEntry
	next      chan *tsdb.Point
	quit      chan bool
}

// NewPool creates a pool of plugin processes corresponding to programs
// held in the given directory. Pool automatically starts/terminates processes
// in response to directory events.
//
// The pool has bounded process count, see MaxProc. An attempt to create
// additional process is logged and ignored.
func NewPool(path string) *Pool {
	ch := make(chan *tsdb.Point, MaxQueue)
	pool := &Pool{
		C:         ch,
		directory: config.WatchDirectory(path),
		byPath:    make(map[string]*directoryEntry),
		next:      ch,
		quit:      make(chan bool),
	}
	statQueue.Set("", expvar.Func(func() interface{} {
		return len(ch)
	}))
	go pool.loop()
	return pool
}

// Kill terminates the process pool. If returns after the last running process
// is reaped.
func (pool *Pool) Kill() {
	pool.quit <- true
	<-pool.quit
}

func (pool *Pool) loop() {
	defer close(pool.quit)
	defer pool.directory.Stop()
	for {
		select {
		case event := <-pool.directory.C:
			if Debug != nil {
				Debug.Printf("pool: directory update, event=%v", event)
			}
			entry := pool.byPath[event.Target]
			if entry == nil {
				if event.IsCreate {
					pool.add(event.Target)
				}
			} else {
				entry.event <- &event
				if event.IsRemove {
					pool.del(event.Target)
				}
			}
		case <-pool.quit:
			if Debug != nil {
				Debug.Printf("pool: got quit request")
			}
			for _, entry := range pool.byPath {
				entry.Kill()
			}
			return
		}
	}
}

// add adds a directory entry to the pool.
func (pool *Pool) add(path string) {
	if max := MaxProc; len(pool.byPath) == max {
		log.Printf("pool: error adding %s: process limit reached (%d)", path, max)
		return
	}
	entry := newEntry(path, pool.next)
	pool.byPath[path] = entry
}

// del deletes the given directory entry from the pool.
func (pool *Pool) del(path string) {
	delete(pool.byPath, path)
}

// directoryEntry represents a program in the directory monitored by Pool.
type directoryEntry struct {
	path         string
	event        chan *config.DirectoryEvent
	w            chan<- *tsdb.Point
	RestartDelay <-chan time.Time
}

func newEntry(path string, w chan<- *tsdb.Point) *directoryEntry {
	entry := &directoryEntry{
		path:  path,
		event: make(chan *config.DirectoryEvent),
		w:     w,
	}
	go entry.mainloop()
	return entry
}

var killRequest = &config.DirectoryEvent{IsRemove: true}

// Kill terminates the process corresponding to the directory entry. It returns
// after the process has been reaped.
func (entry *directoryEntry) Kill() {
	entry.event <- killRequest
	<-entry.event
}

func (entry *directoryEntry) mainloop() {
	defer close(entry.event)
	process := startProcess(entry.path, entry.w)
	for {
		select {
		case event := <-entry.event:
			switch {
			default:
				log.Panicln("unexpected event:", event)
			case event.IsModify:
				if entry.RestartDelay == nil {
					process.Printf("kill (file updated)")
					process.Kill()
					<-process.Exit
				}
				entry.RestartDelay = nil // cancel the restart
				process = startProcess(entry.path, entry.w)
			case event.IsRemove:
				if entry.RestartDelay == nil {
					if event != killRequest {
						process.Printf("kill (file deleted)")
					}
					process.Kill()
					<-process.Exit
				}
				return
			}
		case err := <-process.Exit:
			entry.RestartDelay = restart(process, err)
		case <-entry.RestartDelay:
			entry.RestartDelay = nil
			process = startProcess(entry.path, entry.w)
		}
	}
}
