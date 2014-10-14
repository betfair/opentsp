# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

.PHONY: build clean install

build:
	go install opentsp.org/cmd/tsp-forwarder/...
	go install opentsp.org/cmd/tsp-aggregator/...
	go install opentsp.org/cmd/tsp-poller/...
	go vet opentsp.org/cmd/tsp-forwarder/...
	go vet opentsp.org/cmd/tsp-aggregator/...
	go vet opentsp.org/cmd/tsp-poller/...
	go test opentsp.org/cmd/tsp-forwarder/...
	go test opentsp.org/cmd/tsp-aggregator/...
	go test opentsp.org/cmd/tsp-poller/...

clean:
	rm -rf pkg
	go clean -i ./src/...

install:
	install -m 755 -d $(DESTDIR)/etc/init.d
	install -m 755 dist/forwarder.init $(DESTDIR)/etc/init.d/tsp
	install -m 755 dist/aggregator.init $(DESTDIR)/etc/init.d/tsp-aggregator
	install -m 755 dist/poller.init $(DESTDIR)/etc/init.d/tsp-poller
	install -m 755 -d $(DESTDIR)/etc/sysconfig
	install -m 755 dist/forwarder.sysconfig $(DESTDIR)/etc/sysconfig/tsp
	install -m 755 dist/aggregator.sysconfig $(DESTDIR)/etc/sysconfig/tsp-aggregator
	install -m 755 dist/poller.sysconfig $(DESTDIR)/etc/sysconfig/tsp-poller
	install -m 755 -d $(DESTDIR)/etc/logrotate.d
	install -m 644 dist/forwarder.logrotate $(DESTDIR)/etc/logrotate.d/tsp
	install -m 644 dist/aggregator.logrotate $(DESTDIR)/etc/logrotate.d/tsp-aggregator
	install -m 644 dist/poller.logrotate $(DESTDIR)/etc/logrotate.d/tsp-poller
	install -m 755 -d $(DESTDIR)/etc/tsp/collect.d
	install -m 755 -d $(DESTDIR)/etc/tsp-aggregator/collect.d
	install -m 755 dist/site.sh $(DESTDIR)/etc/tsp-aggregator/collect.d
	install -m 755 -d $(DESTDIR)/etc/tsp-poller/collect.d
	install -m 755 -d $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/collect-site $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/tsp-forwarder $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/tsp-aggregator $(DESTDIR)/usr/bin
	install -m 755 $(GOPATH)/bin/tsp-poller $(DESTDIR)/usr/bin
	install -m 755 -d $(DESTDIR)/usr/share/man/man8
	install -m 755 dist/forwarder.8 $(DESTDIR)/usr/share/man/man8/tsp-forwarder.8
	install -m 755 dist/aggregator.8 $(DESTDIR)/usr/share/man/man8/tsp-aggregator.8
	install -m 755 dist/poller.8 $(DESTDIR)/usr/share/man/man8/tsp-poller.8
	install -m 755 -d $(DESTDIR)/var/log/tsp
