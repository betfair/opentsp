# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

VERSION ?= $(shell echo 0.$$(date +%s)$$USER)

.PHONY: pdf rpm

pdf: tsp-forwarder.pdf tsp-aggregator.pdf tsp-poller.pdf

tsp-forwarder.pdf: forwarder.8
	man -t ./$< | pstopdf -i -o $@

tsp-aggregator.pdf: aggregator.8
	man -t ./$< | pstopdf -i -o $@

tsp-poller.pdf: poller.8
	man -t ./$< | pstopdf -i -o $@

rpm:
	rpmbuild --quiet -bb --define 'version $(VERSION)' service.spec
