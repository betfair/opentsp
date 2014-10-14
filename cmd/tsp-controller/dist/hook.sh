#!/bin/sh
#
# Copyright 2014 The Sporting Exchange Limited. All rights reserved.
# Use of this source code is governed by a free license that can be
# found in the LICENSE file.

# Config validator for tsp-controller(8).
# Install as git pre-commit hook of chef/tsp-controller.

config='config.xml'
url='http://localhost:8084/config/v1/validate'

oldrev=$1
newrev=$2
f=$(mktemp /tmp/tsp-controller-hook.XXXXXXX) || exit
for commit in $(git rev-list ^$oldrev $newrev)
do
	git show $commit:$config >$f || exit
	resp=$(curl <$f -qsS --insecure --max-time 10 --request PUT --data-binary @- $url) || exit
	if [ "$resp" != "ok" ]
	then
		echo >&2 $config': commit '$commit': validation error: '$resp
		rm $f
		exit 1
	fi
done
rm $f
exit 0
