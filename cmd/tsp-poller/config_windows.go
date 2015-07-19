// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

const defaultConfigPath = "poller.json"

var defaultConfig = &Config{
	CollectPath: "collect.d",
	LogPath:     "poller.log",
}
