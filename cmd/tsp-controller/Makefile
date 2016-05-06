# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

export GIT_COMMIT ?= $(shell git rev-parse HEAD)
export GIT_DIRTY ?= $(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
export BUILD_TIME ?= $(shell date +%s)

.PHONY: build install

build:
	go install \
		--ldflags \
		"-X opentsp.org/internal/version.BuildTime=${BUILD_TIME} -X opentsp.org/internal/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}"
	go vet opentsp.org/...
	go test opentsp.org/...

install:
	install -m 755 -d $(DESTDIR)/etc/init.d
	install -m 755 dist/service.init $(DESTDIR)/etc/init.d/tsp-controller
	install -m 755 -d $(DESTDIR)/etc/logrotate.d
	install -m 644 dist/service.logrotate $(DESTDIR)/etc/logrotate.d/tsp-controller
	install -m 755 -d $(DESTDIR)/etc/sysconfig
	install -m 755 dist/service.sysconfig $(DESTDIR)/etc/sysconfig/tsp-controller
	install -m 755 -d $(DESTDIR)/etc/tsp-controller
	install -m 755 dist/service.config $(DESTDIR)/etc/tsp-controller/config
	install -m 755 -d $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/tsp-controller $(DESTDIR)/usr/bin
	install -m 755 -d $(DESTDIR)/usr/share/man/man8
	install -m 755 dist/service.8 $(DESTDIR)/usr/share/man/man8/tsp-controller.8
