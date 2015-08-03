# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

.PHONY: build install

build:
	go install
	go vet opentsp.org/...
	go test opentsp.org/...

install:
	install -m 755 -d $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/collect-netscaler $(DESTDIR)/usr/bin
	install -m 755 -d $(DESTDIR)/usr/share/man/man1
	install -m 755 dist/collector.1 $(DESTDIR)/usr/share/man/man1/collect-netscaler.1
