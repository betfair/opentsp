// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

func init() {
	registerStatFunc("Interface", nic)
}

// TODO: packets_rate
func nic(emit emitFn, r *nitro.ResponseStat) {
	for _, x := range r.Interface {
		post := " interface=" + tsdb.Clean(x.ID)
		emit("nic.bytes direction=In"+post, *x.TotRxBytes)
		emit("nic.bytes direction=Out"+post, *x.TotTxBytes)
		emit("nic.errors direction=In type=Discard"+post, *x.ErrIfInDiscards)
		emit("nic.errors direction=In type=Other"+post, *x.ErrPktRx)
		emit("nic.errors direction=Out type=Discard"+post, *x.NicErrIfOutDiscards)
		emit("nic.errors direction=Out type=Other"+post, *x.ErrPktTx)
		emit("nic.packets direction=In"+post, *x.TotRxPkts)
		emit("nic.packets direction=Out"+post, *x.TotTxPkts)
		// XXX: can't remember why these got commented out. Uncomment and
		// see what blows up.
		// emit("nic.errors direction=In type=Drop"+post, *x.ErrDroppedRxPkts)
		// emit("nic.errors direction=Out type=Drop"+post, *x.ErrDroppedTxPkts)
		// emit("nic.errors direction=Unknown type=Hang"+post, *x.ErrLinkHangs)
		// emit("nic.errors direction=Out type=Stall"+post, *x.NicTxStalls)
		// emit("nic.errors direction=In type=Stall"+post, *x.NicRxStalls)
		// emit("nic.errors direction=NotApplicable type=Disable"+post, *x.NicErrDisables)
		// emit("nic.errors direction=NotApplicable type=DuplexMismatch"+post, *x.ErrDuplexMismatch)
		// emit("nic.errors direction=NotApplicable type=Mute"+post, *x.ErrNicMuted)
	}
}
