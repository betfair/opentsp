// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package config

import (
	"path/filepath"
	"strings"
)

func isDotfile(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}
