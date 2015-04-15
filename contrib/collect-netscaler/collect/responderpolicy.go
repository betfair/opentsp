// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

func init() {
	registerStatFunc("responderpolicy", responderPolicy)
}

func responderPolicy(emit emitFn, r *nitro.ResponseStat) {
	for _, p := range r.ResponderPolicy {
		post := " name=" + tsdb.Clean(*p.Name)
		emit("responderpolicy.HitsRate"+post, *p.HitsRate)
		emit("responderpolicy.UndefHitsRate"+post, *p.UndefHitsRate)
	}
}
