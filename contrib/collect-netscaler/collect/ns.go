// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
)

func init() {
	registerStatFunc("ns", ns)
}

func ns(emit emitFn, r *nitro.ResponseStat) {
	x := r.NS
	emit("cpu.percent type=Mgmt", int(*x.MgmtCPUUsagePcnt))
	emit("cpu.percent type=Packet", int(*x.PktCPUUsagePcnt))
	emit("memory.percent", *x.MemUsagePcnt)
}
