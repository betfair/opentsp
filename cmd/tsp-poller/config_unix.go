// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package main

import "opentsp.org/internal/config"

const defaultConfigPath = "/etc/tsp-poller/config"

var defaultConfig = &Config{
	CollectPath: "/etc/tsp-poller/collect.d",
	LogPath:     "/var/log/tsp/poller.log",
}

func init() {
	go func() {
		config.WaitSIGHUP()
		restartCause <- "received reload signal"
	}()
}
