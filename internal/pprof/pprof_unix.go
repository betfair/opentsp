// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

/*
Package pprof configures export of runtime/pprof data.

Usage:

    import _ "go-nfr/pprof"

The creates a unix domain socket that will serve a HTTP server that exports
pprof data in a fashion analogous to the net/pprof package.

The server listens at /tmp/.go_pid<N>.
*/
package pprof

import (
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
)

func init() {
	s := &server{
		path: listenPath(),
	}
	go s.serve()
}

func listenPath() string {
	dir := os.TempDir()
	file := fmt.Sprintf(".go_pid%d", os.Getpid())
	return filepath.Join(dir, file)
}

// server serves net/http/pprof over a socket file.
type server struct {
	path string
}

func (s *server) serve() {
	// Ensure existing socket is not listening.
	conn, err := net.Dial("unix", s.path)
	if err != nil {
		os.Remove(s.path)
	} else {
		log.Printf("nfr: pprof: %s: already listening", s.path)
		conn.Close()
		return
	}

	// Export pprof.
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))

	// Export /proc.
	fs := http.FileServer(http.Dir("/proc/self"))
	mux.Handle("/debug/proc/", http.StripPrefix("/debug/proc/", fs))

	// Export debugging vars.
	mux.Handle("/debug/vars", http.HandlerFunc(expvarHandler))

	// Serve the handlers.
	l, err := net.Listen("unix", s.path)
	if err != nil {
		log.Printf("nfr: pprof: %s", err)
		return
	}
	os.Chmod(s.path, 0666)
	server := &http.Server{
		Handler: mux,
	}
	err = server.Serve(l)
	log.Printf("nfr: pprof: server exited: %s", err)
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
