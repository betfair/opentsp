// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package collect

import (
	"os/exec"
	"syscall"
)

// statusReschedule is the exit code used to request reschedule.
const statusReschedule = 13

func reschedule(err error) bool {
	if err, ok := err.(*exec.ExitError); ok {
		status := err.Sys().(syscall.WaitStatus)
		return status.Exited() && status.ExitStatus() == statusReschedule
	}
	return false
}
