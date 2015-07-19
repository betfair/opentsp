# Copyright 2015 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

.PHONY: build install

build:
	go install
	go vet
	go test

install:
	install -m 755 -d $(DESTDIR)/etc/init.d
	install -m 755 dist/service.init $(DESTDIR)/etc/init.d/tsp-aggregator
	install -m 755 -d $(DESTDIR)/etc/sysconfig
	install -m 755 dist/service.sysconfig $(DESTDIR)/etc/sysconfig/tsp-aggregator
	install -m 755 -d $(DESTDIR)/etc/logrotate.d
	install -m 644 dist/service.logrotate $(DESTDIR)/etc/logrotate.d/tsp-aggregator
	install -m 755 -d $(DESTDIR)/etc/tsp-aggregator
	install -m 755 -d $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/tsp-aggregator $(DESTDIR)/usr/bin
	install -m 755 -d $(DESTDIR)/usr/share/man/man8
	install -m 755 dist/service.8 $(DESTDIR)/usr/share/man/man8/tsp-aggregator.8
	install -m 755 -d $(DESTDIR)/var/log/tsp
