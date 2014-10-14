// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"os"
	"syscall"
)

// BUG(masiulaniecj): On Windows, plugin rescheduling on exit code 13 is not supported.
func reschedule(_ error) bool { return false }

func sysProcAttr() *syscall.SysProcAttr {
	return nil
}

// BUG(masiulaniecj): On Windows, plugins must run in single process, i.e. must not
// have child processes themselves.
func kill(process *os.Process) error {
	return process.Kill()
}
