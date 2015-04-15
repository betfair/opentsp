// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

func init() {
	registerConfigFunc("filterpolicy", filterPolicy)
}

func filterPolicy(emit emitFn, r *nitro.ResponseConfig) {
	for _, p := range r.FilterPolicy {
		post := " name=" + tsdb.Clean(p.Name)
		emit("filterpolicy.Hits"+post, *p.Hits)
	}
}
