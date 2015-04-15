// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package collect

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/tsdb"
)

func init() {
	registerStatFunc("service", service)
}

func service(emit emitFn, r *nitro.ResponseStat) {
	for vserver, services := range byVServer(r.Service) {
		var total struct {
			ActiveConn            uint64
			ActiveTransactions    uint64
			CurClntConnections    uint64
			CurReusePool          uint64
			CurSrvrConnections    uint64
			RequestBytesRate      float64
			RequestsRate          float64
			ResponseBytesRate     float64
			ResponsesRate         float64
			StateUp               int
			SurgeCount            uint64
			SvrEstablishedConn    uint64
			SvrNotEstablishedConn uint64
		}

		vserver = strings.Replace(vserver, ".", "-", -1)

		for _, svc := range services {
			pre := "service." + tsdb.Clean(vserver) + "."
			post := " service=" + tsdb.Clean(svc.Name)

			stateUp := 1
			if svc.State != "UP" {
				stateUp = 0
			}

			svrNotEstablishedConn := *svc.CurSrvrConnections - *svc.SvrEstablishedConn

			emit(pre+"CurClntConnections"+post, *svc.CurClntConnections)
			emit(pre+"CurSrvrConnections"+post, *svc.CurSrvrConnections)
			emit(pre+"MaxClients"+post, *svc.MaxClients)
			emit(pre+"RequestBytesRate"+post, *svc.RequestBytesRate)
			emit(pre+"ResponseBytesRate"+post, *svc.ResponseBytesRate)
			emit(pre+"StateUp"+post, stateUp)
			emit(pre+"SurgeCount"+post, *svc.SurgeCount)
			emit(pre+"SvrEstablishedConn"+post, *svc.SvrEstablishedConn)
			emit(pre+"SvrNotEstablishedConn"+post, svrNotEstablishedConn)

			// NB: MaxClients deliberately omitted; it's confusing when aggregated.
			total.CurClntConnections += *svc.CurClntConnections
			total.CurSrvrConnections += *svc.CurSrvrConnections
			total.RequestBytesRate += *svc.RequestBytesRate
			total.ResponseBytesRate += *svc.ResponseBytesRate
			total.StateUp += stateUp
			total.SurgeCount += *svc.SurgeCount
			total.SvrEstablishedConn += *svc.SvrEstablishedConn
			total.SvrNotEstablishedConn += svrNotEstablishedConn

			if svc.ServiceType == "HTTP" {
				activeConn := *svc.SvrEstablishedConn - *svc.CurReusePool

				emit(pre+"ActiveConn"+post, activeConn)
				emit(pre+"ActiveTransactions"+post, *svc.ActiveTransactions)
				emit(pre+"AvgSvrTTFB"+post, *svc.AvgSvrTTFB)
				emit(pre+"CurReusePool"+post, *svc.CurReusePool)
				emit(pre+"RequestsRate"+post, *svc.RequestsRate)
				emit(pre+"ResponsesRate"+post, *svc.ResponsesRate)

				total.ActiveConn += activeConn
				total.ActiveTransactions += *svc.ActiveTransactions
				total.CurReusePool += *svc.CurReusePool
				total.RequestsRate += *svc.RequestsRate
				total.ResponsesRate += *svc.ResponsesRate
			}
		}

		post := " vserver=" + tsdb.Clean(vserver)

		emit("total.service.CurClntConnections"+post, total.CurClntConnections)
		emit("total.service.CurSrvrConnections"+post, total.CurSrvrConnections)
		emit("total.service.RequestBytesRate"+post, total.RequestBytesRate)
		emit("total.service.ResponseBytesRate"+post, total.ResponseBytesRate)
		emit("total.service.StateUp"+post, total.StateUp)
		emit("total.service.SurgeCount"+post, total.SurgeCount)
		emit("total.service.SvrEstablishedConn"+post, total.SvrEstablishedConn)
		emit("total.service.SvrNotEstablishedConn"+post, total.SvrNotEstablishedConn)

		if len(services) > 0 && services[0].ServiceType == "HTTP" {
			emit("total.service.ActiveConn"+post, total.ActiveConn)
			emit("total.service.ActiveTransactions"+post, total.ActiveTransactions)
			emit("total.service.CurReusePool"+post, total.CurReusePool)
			emit("total.service.RequestsRate"+post, total.RequestsRate)
			emit("total.service.ResponsesRate"+post, total.ResponsesRate)
		}
	}
}

// byVServer partitions service list by the VServer they are bound to.
// It also discards VServers that are idle or bindingless.
func byVServer(all []nitro.Service) map[string][]nitro.Service {
	active := lbvservers.RecentlyActive()
	lookup := serviceMap(all)

	m := make(map[string][]nitro.Service, len(active))

	for vserver, services := range serviceBindings.Map() {
		if !active[vserver] || len(services) == 0 {
			continue
		}
		m[vserver] = make([]nitro.Service, 0, len(services))
		for _, name := range services {
			m[vserver] = append(m[vserver], lookup[name])
		}
	}

	return m
}

func serviceMap(services []nitro.Service) map[string]nitro.Service {
	clash := make(map[string]bool)
	m := make(map[string]nitro.Service, len(services))
	for _, s := range services {
		if clash[s.Name] {
			continue
		}
		if _, ok := m[s.Name]; ok {
			log.Printf("nitro data error: duplicate entry for service %s", s.Name)
			clash[s.Name] = true
			delete(m, s.Name)
			continue
		}
		m[s.Name] = s
	}
	return m
}

var serviceBindings = newServiceBindingsPoller()

type currentServiceBindings struct {
	byLBVServer   map[string][]string
	byLBVServerMu sync.Mutex
	running       map[string]*bindingReaderJob
}

func newServiceBindingsPoller() *currentServiceBindings {
	sb := &currentServiceBindings{
		byLBVServer: make(map[string][]string),
		running:     make(map[string]*bindingReaderJob),
	}
	go sb.loop()
	return sb
}

// Map returns a copy of the bindings map.
func (sb *currentServiceBindings) Map() map[string][]string {
	sb.byLBVServerMu.Lock()
	defer sb.byLBVServerMu.Unlock()

	m := make(map[string][]string)
	for vserver, bindings := range sb.byLBVServer {
		m[vserver] = bindings
	}
	return m
}

func (sb *currentServiceBindings) loop() {
	updateChan := make(chan binding)
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			avail := lbvservers.RecentlyActive()
			for id := range avail {
				if sb.running[id] == nil {
					sb.running[id] = newBindingReaderJob(id, updateChan)
				}
			}
			for id := range sb.running {
				if !avail[id] {
					sb.running[id].Exit <- true
					delete(sb.running, id)
				}
			}

		case b := <-updateChan:
			sb.byLBVServerMu.Lock()
			sb.byLBVServer[b.LBID] = b.ServiceList
			sb.byLBVServerMu.Unlock()
		}
	}
}

type binding struct {
	LBID        string
	ServiceList []string
}

// bindingReaderJob is a background job that keeps LBVServer-Service binding data
// up to date.
type bindingReaderJob struct {
	lbID string
	out  chan<- binding
	Exit chan bool
	rate *time.Ticker
}

// newBindingReaderJob starts a new binding-reading job. The binding readings
// will be sent to the given out channel.
func newBindingReaderJob(lbID string, out chan<- binding) *bindingReaderJob {
	job := &bindingReaderJob{
		lbID: lbID,
		out:  out,
		Exit: make(chan bool, 1),
	}
	go job.mainloop()
	return job
}

// getBoundServices returns names of services bound to the given LBVServer.
func getBoundServices(lbID string) ([]string, error) {
	resp, err := Client.Config.Get("lbvserver_service_binding/" + lbID)
	if err != nil {
		return nil, err
	}
	got := resp.LBVServerServiceBinding
	names := make([]string, 0, len(got))
	for _, s := range got {
		names = append(names, s.ServiceName)
	}
	return names, nil
}

func (job *bindingReaderJob) mainloop() {
	gotOK := 0
	for {
		job.sleep(gotOK)
		svcList, err := getBoundServices(job.lbID)
		if err != nil {
			log.Print(err)
			continue
		}
		gotOK++
		select {
		case job.out <- binding{job.lbID, svcList}:
			// ok
		case <-job.Exit:
			if job.rate != nil {
				job.rate.Stop()
			}
			return
		}
	}
}

// bindingReadInterval sets a per-LBVServer rate limit on calls to the Nitro API
// requesting the LBVServer's service bindings. Bindings retrieval is a background
// job that enables Loop to make cheaper Nitro calls, and thus meet the performance
// requirement of ~5s collection cycle.
const bindingReadInterval = 10 * time.Minute

// bindingReadRateLimit is an upper bound on the rate of calls to the Nitro API
// requesting the VServer-Service bindings. Without the limit, Nitro API crashes
// and takes up to a minute to recover causing gaps in data.
var bindingReadRateLimit = time.NewTicker(100 * time.Millisecond)

// sleep controls the delay between binding read requests.
func (job *bindingReaderJob) sleep(got int) {
	switch got {
	case 0:
		// No binding data available yet; request it as fast as permitted.
		<-bindingReadRateLimit.C
	case 1:
		// Got first binding. Delay second read to avoid thundering herds.
		splay(bindingReadInterval)
		job.rate = time.NewTicker(bindingReadInterval)
	default:
		// Regular binding request: no special delay needed. Only need to
		// respect the rate limits.
		<-job.rate.C
		<-bindingReadRateLimit.C
	}
}

func splay(n time.Duration) {
	time.Sleep(time.Duration(rand.Int63n(n.Nanoseconds())))
}
