// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package main

import (
	"opentsp.org/contrib/collect-netscaler/collect"
	"opentsp.org/contrib/collect-netscaler/config"

	_ "opentsp.org/internal/pprof"
)

func main() {
	collect.Loop(tsdbChan, *config.Interval)
}
