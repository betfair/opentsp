// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package control handles config requests from collect-jmx.
//
// The lookup requests use the simple key host, or the composite (host,
// process).
//
// The host-specific settings are always served. These setting may be
// extended using the process-specific settings, if found.
package control

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"reflect"

	"opentsp.org/cmd/tsp-controller/config"
	"opentsp.org/cmd/tsp-controller/control"
)

const (
	defaultInterval = 10000 // default mbean poll interval
)

func init() {
	control.Register(Update, Handle)
}

// Update decodes the collect-jmx parts of the config.
func Update(config *config.Config) error {
	if err := updateHost(config); err != nil {
		return err
	}
	if err := updateGlobal(config); err != nil {
		return err
	}
	return nil
}

// updateGlobal extends the provided common config with global QueryGroup elements.
func updateGlobal(config_ *config.Config) error {
	ns := make(config.Namespace)
	for _, elem := range config_.Extra {
		if elem.Value != nil {
			continue
		}
		g := new(QueryGroup)
		if err := xml.Unmarshal(elem.Raw, g); err != nil {
			continue
		}
		if err := g.validate(config_); err != nil {
			return err
		}
		if err := ns.Add(g.ID); err != nil {
			return fmt.Errorf("querygroup %s: %v", g.ID, err)
		}
		elem.Value = g
	}
	return nil
}

// updateHost extends the provided common config with host-level Process elements.
func updateHost(config_ *config.Config) error {
	global := make(config.Namespace)
	before := make(config.Namespace)
	for id := range config_.Hosts.NS {
		before[id] = true
	}
	// Decode <process> elements.
	err := config_.UpdateHost(func(hostID string, elem *config.Element) error {
		p := new(Process)
		if err := xml.Unmarshal(elem.Raw, p); err != nil {
			return nil
		}
		if err := p.validate(); err != nil {
			return fmt.Errorf("host %s: %v", hostID, err)
		}
		config_.Hosts.NS[p.ID] = true
		global[p.ID] = true
		elem.Value = p
		return nil
	})
	if err != nil {
		return err
	}
	// Ensure no naming ambiguity exists.
	count := make(map[string]int)
	for id := range global {
		count[id]++
	}
	for id := range before {
		count[id]++
	}
	for id := range count {
		if count[id] > 1 {
			return fmt.Errorf("identifier redeclared: %v", id)
		}
	}
	return nil
}

// A QueryGroup corresponds to a <querygroup> configuration element.
type QueryGroup struct {
	ID      string
	Targets []string
	Query   []*Query
}

var _ xml.Unmarshaler = &QueryGroup{}

func (g *QueryGroup) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v struct {
		XMLName xml.Name  `xml:"querygroup"`
		ID      string    `xml:"id,attr"`
		Targets *string   `xml:"targets,attr"`
		Query   []*Query  `xml:"query"`
		Any     *xml.Name `xml:",any"`
	}
	if err := dec.DecodeElement(&v, &start); err != nil {
		return err
	}
	if any := v.Any; any != nil {
		return fmt.Errorf("querygroup %s: invalid element: %s", g.ID, any.Local)
	}
	var targets []string
	switch p := v.Targets; {
	case p != nil && *p != "":
		targets = strings.Split(*p, ",")
	case p != nil:
		targets = []string{}
	}
	*g = QueryGroup{
		ID:      v.ID,
		Targets: targets,
		Query:   v.Query,
	}
	return nil
}

func (g *QueryGroup) Match(ids ...string) bool {
	for _, want := range g.Targets {
		for _, got := range ids {
			if got == want {
				return true
			}
		}
	}
	return false
}

func (g *QueryGroup) validate(config *config.Config) error {
	if g.ID == "" {
		return fmt.Errorf("missing querygroup attribute: id")
	}
	if g.Targets == nil {
		return fmt.Errorf("querygroup %s: missing attribute: targets", g.ID)
	}
	seen := make(map[interface{}]bool)
	for _, targetID := range g.Targets {
		if !config.Hosts.NS[targetID] {
			return fmt.Errorf("undefined: %s", targetID)
		}
		for _, host := range config.Hosts.All {
			for _, id := range host.Tags {
				if id == targetID {
					if seen[host] {
						return fmt.Errorf("host %s: query redeclared", host.ID)
					}
					seen[host] = true
				}
			}
			for _, elem := range host.Extra {
				process, ok := elem.Value.(*Process)
				if !ok {
					continue
				}
				if process.ID == targetID {
					if seen[process] {
						return fmt.Errorf("process %s: query redeclared", host.ID)
					}
					seen[process] = true
				}
			}
		}
	}
	for _, query := range g.Query {
		if err := query.validate(); err != nil {
			return fmt.Errorf("querygroup %s: %v", g.ID, err)
		}
	}
	return nil
}

// A Process corresponds to a <process> configuration element.
type Process struct {
	ID string
}

var _ = xml.Unmarshaler(&Process{})

func (p *Process) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var v struct {
		XMLName xml.Name  `xml:"process"`
		ID      string    `xml:"id,attr"`
		Any     *xml.Name `xml:",any"`
	}
	if err := dec.DecodeElement(&v, &start); err != nil {
		return err
	}
	if any := v.Any; any != nil {
		return fmt.Errorf("process %s: invalid element: %s", p.ID, any.Local)
	}
	*p = Process{ID: v.ID}
	return nil
}

func (p *Process) validate() error {
	if p.ID == "" {
		return fmt.Errorf("missing process attribute: id")
	}
	// XXX: check namespace conflict
	return nil
}

// Handle registers the HTTP handlers.
func Handle(config *config.Config) error {
	handler := &handler{config}
	http.Handle("/control/v1/collect-jmx", handler)
	http.Handle("/config/betex", handler) // legacy
	http.Handle("/config/jmx", handler)   // legacy
	return nil
}

// Key represents the configuration lookup request.
type Key struct {
	Program, Host, Process string
}

func parseKey(url *url.URL) (*Key, error) {
	query := url.Query()
	host := query.Get("host")
	if host == "" {
		return nil, fmt.Errorf("missing query parameter: host")
	}
	query.Del("host")
	process := query.Get("process")
	query.Del("process")
	for _, key := range []string{"version", "wibble"} {
		query.Del(key)
	}
	for key := range query {
		return nil, fmt.Errorf("unexpected query parameter: %s", key)
	}
	return &Key{"collect-jmx", host, process}, nil
}

// View corresponds to collect-jmx configuration file.
type View struct {
	Period  int      `json:"period"`
	Objects []*Query `json:"objects"`
}

func (v *View) appendObject(nq *Query) {
	for _,item := range v.Objects {
		if reflect.DeepEqual(nq, item) {
			return
		}
	}
	v.Objects = append(v.Objects, nq)
}

// An Object is the JSON structure describing a Query.
type Object struct {
	// Attributes corresponds to Query's Attributes field.
	Attributes []*Attribute `json:"attributes,omitempty"`

	// KPIs corresponds to Query's KPI field.
	KPIs []*KPI `json:"kpis"`

	// MetricName corresponds to Query's ID field.
	MetricName string `json:"metricName"`

	// ObjectName corresponds to Query's On field.
	ObjectName string `json:"objectName"`
}

// Attribute represent a single mbean attribute.
type Attribute struct {
	Name string `json:"attributeName"`
}

// A KPI is the JSON structure of a KPIQuery.
type KPI struct {
	Name          string `xml:"name,attr" json:"name"`
	ID            string `xml:"id,attr" json:"id"`
	CollectRanges bool   `xml:"collectRanges,attr" json:"collectRanges"`
}

func (kpi *KPI) validate() error {
	if kpi.ID == "" {
		return fmt.Errorf("missing kpi attribute: id")
	}
	if kpi.Name == "" {
		return fmt.Errorf("kpi %s: missing attribute: name", kpi.ID)
	}
	return nil
}

type handler struct {
	config *config.Config
}

func (h *handler) View(key *Key) (*View, error) {
	// Lookup the host.
	host, err := h.config.Host(key.Host)
	if err != nil {
		return nil, err
	}
	view := new(View)
	// Add host-level queries.
	view.Objects = hostQueries(host, h.config)
	// Add process-level queries.
	for _, query := range processQueries(host, h.config, key.Process) {
		view.appendObject(query)
	}
	// Set poll interval.
	view.Period = defaultInterval
	return view, nil
}

func hostQueries(host *config.Host, config *config.Config) []*Query {
	var found []*Query
	for _, elem := range config.Extra {
		group := elem.Value.(*QueryGroup)
		if group.Match(host.Tags...) {
			found = append(found, group.Query...)
		}
	}
	return found
}

func processQueries(host *config.Host, config *config.Config, processID string) []*Query {
	var found []*Query
	for _, elem := range config.Extra {
		group := elem.Value.(*QueryGroup)
		if group.Match(processID) {
			found = append(found, group.Query...)
		}
	}
	return found
}

type internalError struct{ error }

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key, err := parseKey(req.URL)
	if err != nil {
		log.Printf("%v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	view, err := h.View(key)
	if err != nil {
		log.Printf("%s: %v", key.Program, err)
		if _, ok := err.(*internalError); ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	control.Marshal(w, view)
}

// Query represents a <query> element in the config file.
type Query struct {
	ID         string
	On         string
	Attributes []string
	KPI        []*KPI
}

func (q *Query) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var e struct {
		ID         string `xml:"id,attr"`
		On         string `xml:"on,attr"`
		Attributes string `xml:"attributes,attr"`
		KPI        []*KPI `xml:"kpi"`
	}
	if err := dec.DecodeElement(&e, &start); err != nil {
		return err
	}
	var attributes []string
	if e.Attributes != "" {
		attributes = strings.Split(e.Attributes, ",")
	}
	*q = Query{e.ID, e.On, attributes, e.KPI}
	return nil
}

func (q *Query) validate() error {
	if q.ID == "" {
		return fmt.Errorf("missing query attribute: id")
	}
	if q.On == "" {
		return fmt.Errorf("query %s: missing attribute: on", q.ID)
	}
	seen := make(map[string]bool)
	for _, attr := range q.Attributes {
		if seen[attr] {
			return fmt.Errorf("query %s: attribute redeclared: %s", q.ID, attr)
		}
		seen[attr] = true
	}
	for _, kpi := range q.KPI {
		if err := kpi.validate(); err != nil {
			return fmt.Errorf("query %s: %v", q.ID, err)
		}
	}
	return nil
}

func (q *Query) MarshalJSON() ([]byte, error) {
	object := &Object{
		Attributes: q.attributes(),
		KPIs:       q.kpis(),
		ObjectName: q.On,
		MetricName: q.ID,
	}
	return json.Marshal(object)
}

func (q *Query) attributes() []*Attribute {
	var attrs []*Attribute
	for _, name := range q.Attributes {
		attrs = append(attrs, &Attribute{
			Name: name,
		})
	}
	return attrs
}

func (q *Query) kpis() []*KPI {
	kpis := make([]*KPI, 0, len(q.KPI))
	for _, k := range q.KPI {
		kpis = append(kpis, &KPI{
			Name:          k.Name,
			ID:            k.ID,
			CollectRanges: k.CollectRanges,
		})
	}
	return kpis
}
