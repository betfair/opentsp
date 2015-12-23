// Copyright 2015 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package version reports build metadata for the currently running binary.
package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

// These vars get bound at build time using the --ldflags mechanism.
var (
	Version   = "0.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func String() string {
	built := BuildTime
	i, err := strconv.ParseInt(BuildTime, 10, 64)
	if err == nil {
		t := time.Unix(i, 0)
		built = t.Format(time.RFC3339)
	}
	return fmt.Sprintf("%s (git=%s) (built %s) using %s\n", Version, GitCommit, built, runtime.Version())
}
