// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package config

import (
	"os"
	"os/signal"
	"syscall"
)

var defaultByProgram = map[string]*Config{
	"forwarder": {
		FilePath:    "/etc/tsp/config",
		CollectPath: "/etc/tsp/collect.d",
		LogPath:     "/var/log/tsp/forwarder.log",
	},
	"aggregator": {
		FilePath:    "/etc/tsp-aggregator/config",
		CollectPath: "/etc/tsp-aggregator/collect.d",
		LogPath:     "/var/log/tsp/aggregator.log",
	},
	"poller": {
		FilePath:    "/etc/tsp-poller/config",
		CollectPath: "/etc/tsp-poller/collect.d",
		LogPath:     "/var/log/tsp/poller.log",
	},
}

func sighupHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Signal(syscall.SIGHUP))
	<-ch
	reload <- "received reload signal"
}

func init() {
	go sighupHandler()
}
