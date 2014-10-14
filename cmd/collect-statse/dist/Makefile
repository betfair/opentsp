# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

VERSION ?= $(shell echo 0.$$(date +%s)$$USER)

.PHONY: rpm pdf

rpm:
	rpmbuild --quiet -bb --define 'version $(VERSION)' collector.spec

pdf: \
	collect-statse.pdf \
	statse.pdf

collect-statse.pdf: collector.1
	man -t ./$< | pstopdf -i -o $@

statse.pdf: protocol.5
	man -t ./$< | pstopdf -i -o $@
