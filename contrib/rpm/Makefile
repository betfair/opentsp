# Copyright 2015 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

VERSION ?= $(shell echo 0.$$(date +%s)$$USER)

.PHONY: rpm

rpm:
	rpmbuild -bb --define 'version $(VERSION)' tsp.spec
