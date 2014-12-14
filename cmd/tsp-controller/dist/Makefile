# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

VERSION ?= $(shell echo 0.$$(date +%s)$$USER)

.PHONY: rpm pdf

rpm:
	rpmbuild -bb --define 'version $(VERSION)' service.spec

pdf: tsp-controller.pdf

tsp-controller.pdf: service.8
	man -t ./$< | pstopdf -i -o $@
