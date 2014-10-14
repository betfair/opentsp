// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package config

var defaultByProgram = map[string]*Config{
	"forwarder": {
		FilePath:    "forwarder.json",
		CollectPath: "collect.d",
		LogPath:     "forwarder.log",
	},
	"aggregator": {
		FilePath:    "aggregator.json",
		CollectPath: "collect.d",
		LogPath:     "aggregator.log",
	},
	"poller": {
		FilePath:    "poller.json",
		CollectPath: "collect.d",
		LogPath:     "poller.log",
	},
}
