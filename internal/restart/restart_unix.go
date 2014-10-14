// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package restart

import (
	"log"
	"os"
	"syscall"
	"time"
)

func Do() {
	for {
		argv0 := os.Args[0]
		closeFD()
		err := syscall.Exec(argv0, os.Args, syscall.Environ())
		if err != nil {
			log.Printf("restart error: %s: %v", argv0, err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Panicf("unreachable")
	}
}

// maxFD is the maximum file number closed by closeFD. It aims to include
// low-numbered system files, e.g. /dev/urandom.
const maxFD = 128

// closeFD is required for systems that don't support O_CLOEXEC open(2) flag, for
// example Linux < 2.6.23.
func closeFD() {
	for fd := 3; fd < maxFD; fd++ {
		syscall.CloseOnExec(fd)
	}
}
