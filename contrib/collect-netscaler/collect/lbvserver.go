// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"strings"
	"sync"
	"time"

	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

func init() {
	registerStatFunc("lbvserver", lbvserver)
}

func lbvserver(emit emitFn, r *nitro.ResponseStat) {
	lbvservers.Snapshot(r.LBVServer...)
	active := lbvservers.RecentlyActive()

	for _, vs := range r.LBVServer {
		if !active[vs.Name] {
			continue
		}
		name := strings.Replace(vs.Name, ".", "-", -1)
		pre := "vserver."
		post := " vserver=" + tsdb.Clean(name)

		stateUp := 1
		if vs.State != "UP" {
			stateUp = 0
		}

		emit(pre+"CurClntConnections"+post, *vs.CurClntConnections)
		emit(pre+"CurSrvrConnections"+post, *vs.CurSrvrConnections)
		emit(pre+"Health"+post, *vs.Health)
		emit(pre+"RequestBytesRate"+post, *vs.RequestBytesRate)
		emit(pre+"ResponseBytesRate"+post, *vs.ResponseBytesRate)
		emit(pre+"SpilloverThreshold"+post, *vs.SpilloverThreshold)
		emit(pre+"Spillovers"+post, *vs.Spillovers)
		emit(pre+"StateUp"+post, stateUp)

		if vs.Type == "HTTP" {
			emit(pre+"RequestsRate"+post, *vs.RequestsRate)
			emit(pre+"ResponsesRate"+post, *vs.ResponsesRate)
		}
	}
}

var lbvservers = newLBVServerLog()

// lbvserverLog maintains a subset of all VServer names configured on the
// device. The subset includes those VServers that have been observed to
// receive traffic within the past 24 hours.
//
// Restricting collection to recently accessed VServers prevents excessive
// response latency caused by defunct VServers configured on the device.
type lbvserverLog struct {
	byName map[string]*lbvserverStatus
	mu     sync.Mutex
}

type lbvserverStatus struct {
	seen          time.Time
	totalPktsSent uint64
}

func newLBVServerLog() *lbvserverLog {
	return &lbvserverLog{
		byName: make(map[string]*lbvserverStatus),
	}
}

// Snapshot merges most recent information about the configured VServers.
func (l *lbvserverLog) Snapshot(all ...nitro.LBVServer) {
	for _, vs := range all {
		l.merge(vs)
	}
}

func (l *lbvserverLog) merge(vs nitro.LBVServer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	status := l.byName[vs.Name]
	if status == nil {
		status = &lbvserverStatus{
			totalPktsSent: *vs.TotalPktsSent,
		}
		l.byName[vs.Name] = status
		return
	}
	if tps := *vs.TotalPktsSent; tps != status.totalPktsSent {
		status.seen = time.Now()
		status.totalPktsSent = tps
	}
}

func (l *lbvserverLog) RecentlyActive() map[string]bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	active := make(map[string]bool)
	for name, status := range l.byName {
		if time.Since(status.seen) < 24*time.Hour {
			active[name] = true
		}
	}
	return active
}
