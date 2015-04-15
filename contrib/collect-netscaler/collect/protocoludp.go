// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
)

func init() {
	registerStatFunc("protocoludp", protocolUDP)
}

func protocolUDP(emit emitFn, r *nitro.ResponseStat) {
	x := r.ProtocolUDP

	emit("protocol.udp.BadChecksum", *x.BadChecksum)
	emit("protocol.udp.CurRateThreshold", *x.CurRateThreshold)
	emit("protocol.udp.CurRateThresholdExceeds", *x.CurRateThresholdExceeds)
	emit("protocol.udp.RxBytesRate", *x.RxBytesRate)
	emit("protocol.udp.RxPktsRate", *x.RxPktsRate)
	emit("protocol.udp.TotUnknownSvcPkts", *x.TotUnknownSvcPkts)
	emit("protocol.udp.TxBytesRate", *x.TxBytesRate)
	emit("protocol.udp.TxPktsRate", *x.TxPktsRate)
}
