# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

export GIT_COMMIT ?= $(shell git rev-parse HEAD)
export GIT_DIRTY ?= $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
export BUILD_TIME ?= $(shell date +%s)

export CGO_CFLAGS ?= $(shell pkg-config --cflags libzmq)
export CGO_LDFLAGS ?= $(shell pkg-config --libs libzmq)

.PHONY: build install

build:
	go install -x -v \
		--ldflags \
		"-X opentsp.org/internal/version.BuildTime=${BUILD_TIME} -X opentsp.org/internal/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} "
	go vet opentsp.org/...
	go test opentsp.org/...

install:
	install -m 755 -d $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/collect-statse $(DESTDIR)/usr/bin
	install -m 755 -d $(DESTDIR)/usr/share/man/man5
	install -m 755 dist/protocol.5 $(DESTDIR)/usr/share/man/man5/statse.5
	install -m 755 -d $(DESTDIR)/usr/share/man/man1
	install -m 755 dist/collector.1 $(DESTDIR)/usr/share/man/man1/collect-statse.1
