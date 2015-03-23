// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"opentsp.org/contrib/collect-netscaler/nitro"
)

func init() {
	registerStatFunc("protocolhttp", protocolHTTP)
}

func protocolHTTP(emit emitFn, r *nitro.ResponseStat) {
	x := r.ProtocolHTTP
	emit("protocol.http.request.bytes", *x.HTTPTotRxRequestBytes)
	emit("protocol.http.request.errors type=HeaderTooLong", *x.HTTPErrIncompleteRequests)
	emit("protocol.http.request.received type=GET", *x.HTTPTotGets)
	emit("protocol.http.request.received type=Other", *x.HTTPTotOthers)
	emit("protocol.http.request.received type=POST", *x.HTTPTotPosts)
	emit("protocol.http.response.bytes", *x.HTTPTotTxResponseBytes)
	emit("protocol.http.response.errors type=5yz", *x.HTTPErrServerBusy)
	emit("protocol.http.response.errors type=HeaderTooLong", *x.HTTPErrIncompleteResponses)
	emit("protocol.http.response.sent", *x.HTTPTotResponses)
}
