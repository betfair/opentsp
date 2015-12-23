// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package config handles the main config file.
package config

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"opentsp.org/cmd/tsp-controller/config/network"
	"opentsp.org/internal/version"
)

// maxSize limits the size of the config stream.
const maxSize = 10 << 20 // 10 MB

var Debug *log.Logger

var (
	FilePath    = flag.String("f", DefaultFilePath, "configuration file")
	TestMode    = flag.Bool("t", false, "configuration test")
	VerboseMode = flag.Bool("v", false, "verbose mode")
	ListenAddr  = flag.String("l", ":8084", "listen address")
	VersionMode = flag.Bool("version", false, "echo version and exit")
)

var modules []func(*Config) error

// Register registers a function that updates a Config by decoding the ExtraRaw fields.
func Register(fn func(*Config) error) {
	modules = append(modules, fn)
}

// Loaded contains the loaded configuration.
var Loaded *Config

// Config represents controller's entire configuration, i.e. the <config> block.
type Config struct {
	Hosts   *Hosts          `xml:"hostgroup"`
	Network *network.Config `xml:"network"`
	Filter  *Filter         `xml:"filter"`
	Extra   []*Element      `xml:",any"`
}

// UpdateFunc is registered by extension module to decode custom Host-level
// elements. It is passed the Host.ID and an Element to decode. The decoded
// information must be stored in Element.Value. Any error encountered should
// be returned.
type UpdateFunc func(string, *Element) error

// UpdateHost updates config to include extra Host-level elements.
func (c *Config) UpdateHost(fn UpdateFunc) error {
	for _, host := range c.Hosts.All {
		for _, elem := range host.Extra {
			if elem.Value != nil {
				continue
			}
			if err := fn(host.ID, elem); err != nil {
				return err
			}
		}
	}
	return nil
}

// Host returns a host with the given id.
func (c *Config) Host(id string) (*Host, error) {
	var (
		host         *Host
		inexact      *Host
		inexactCount = 0
	)
	for _, registered := range c.Hosts.All {
		if registered.ID == id {
			host = registered
			break
		}
		if registered.shortID() == id {
			inexact = registered
			inexactCount++
			continue
		}
	}
	// If not exact match was found, use inexact if unambiguous.
	if host == nil && inexactCount == 1 {
		host = inexact
	}
	// If no host was found, invent a config for one anyway.
	if host == nil {
		host = &Host{ID: id}
	}
	// Apply scope restrictions.
	if !c.Network.InScope(host.ID) {
		return nil, fmt.Errorf("scope error: host %s rejected due to scope restriction", host.ID)
	}
	return host, nil
}

// Cluster returns hosts with the given cluster id.
func (c *Config) Cluster(id string) []*Host {
	var found []*Host
	for _, host := range c.Hosts.All {
		if !c.Network.InScope(host.ID) {
			continue
		}
		if host.ClusterID != id {
			continue
		}
		found = append(found, host)
	}
	return found
}

// Decode decodes from r the common configuration settings. It also validates
// that the extended settings are correct, but omits them from the returned config.
func Decode(r io.Reader) (*Config, error) {
	dec := newDecoder(r)
	return dec.Decode()
}

type decoder struct {
	dec *xml.Decoder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		dec: xml.NewDecoder(io.LimitReader(r, maxSize)),
	}
}

func (d *decoder) Decode() (*Config, error) {
	config := Config{
		Hosts: &Hosts{
			NS: make(Namespace),
		},
		Filter: &Filter{},
	}
	if err := d.dec.Decode(&config); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	config.Hosts.tags = nil
	if err := validateFilter(&config); err != nil {
		return nil, err
	}
	if err := validateExtra(&config); err != nil {
		return nil, err
	}
	if err := validateNetwork(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func validateNetwork(config *Config) error {
	if config.Network == nil {
		network, err := network.ReadFile(network.DefaultPath)
		if err != nil {
			return err
		}
		config.Network = network
	}
	return nil
}

func validateFilter(config *Config) error {
	if config.Filter.Path == "" {
		config.Filter.Path = DefaultFilterPath
	}
	return nil
}

func validateExtra(config *Config) error {
	for _, updateFn := range modules {
		if err := updateFn(config); err != nil {
			return err
		}
	}
	// Ensure no unsupported elements remain.
	for _, elem := range config.Extra {
		if elem.Value == nil {
			err := fmt.Errorf("unsupported element: %v", elem.Name)
			return err
		}
	}
	for _, host := range config.Hosts.All {
		for _, elem := range host.Extra {
			if elem.Value == nil {
				err := fmt.Errorf("unsupported element: %v", elem.Name)
				return err
			}
		}
	}
	return nil
}

// Load loads configuration settings from FilePath, storing them in Loaded.
func Load() {
	flag.Parse()
	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}
	if *VersionMode {
		fmt.Println(version.String())
		os.Exit(0)
	}
	if *VerboseMode {
		Debug = log.New(os.Stderr, "debug: ", 0)
	}
	buf, err := ioutil.ReadFile(*FilePath)
	if err != nil {
		log.Fatal(err)
	}
	config, err := Decode(bytes.NewBuffer(buf))
	if err != nil {
		log.Fatal(err)
	}
	if *TestMode {
		os.Exit(0)
	}
	log.Print("start",
		" pid=", os.Getpid(),
	)
	Loaded = config
}

// Host represents a single <host> element.
type Host struct {
	ID        string     `xml:"id,attr"`
	ClusterID string     `xml:"-"`
	Tags      []string   `xml:"-"`
	Extra     []*Element `xml:",any"`
}

func (h *Host) Validate() error {
	if h.ID == "" {
		return fmt.Errorf("missing host attribute: id")
	}
	return nil
}

// shortID shortens "foo.example.com" to just "foo".
func (h *Host) shortID() string {
	label := strings.Split(h.ID, ".")
	if len(label) == 1 {
		return h.ID
	}
	return label[0]
}

func (h *Host) String() string {
	return fmt.Sprintf("%v", *h)
}

// Cluster represents a single <cluster> element.
type Cluster struct {
	ID   string    `xml:"id,attr"`
	Host []*Host   `xml:"host"`
	Any  *xml.Name `xml:",any"`
}

func (cluster *Cluster) Validate() error {
	if cluster.ID == "" {
		return fmt.Errorf("missing cluster attribute: id")
	}
	if any := cluster.Any; any != nil {
		return fmt.Errorf("cluster %s: invalid element: %s", cluster.ID, any.Local)
	}
	return nil
}

// HostGroup represents a single <hostgroup> element.
type HostGroup struct {
	Sub     []*HostGroup `xml:"hostgroup"`
	ID      string       `xml:"id,attr"`
	Cluster []*Cluster   `xml:"cluster"`
	Any     *xml.Name    `xml:",any"`
}

func (group *HostGroup) Validate() error {
	if group.ID == "" {
		return fmt.Errorf("missing hostgroup attribute: id")
	}
	if any := group.Any; any != nil {
		return fmt.Errorf("hostgroup %s: invalid element: %s", group.ID, any.Local)
	}
	return nil
}

// Hosts is a set of all <host> elements.
type Hosts struct {
	All  []*Host
	NS   Namespace
	tags tags
}

func (hs *Hosts) String() string {
	return fmt.Sprintf("%v", hs.All)
}

func (hs *Hosts) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	group := new(HostGroup)
	if err := dec.DecodeElement(group, &start); err != nil {
		return err
	}
	if err := hs.addGroup(group); err != nil {
		return err
	}
	return nil
}

func (hs *Hosts) addGroup(group *HostGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}
	if err := hs.NS.Add(group.ID); err != nil {
		return err
	}
	hs.tags.Push(group.ID)
	defer hs.tags.Pop()
	switch {
	default:
		// ok
	case len(group.Sub) > 0:
		if len(group.Cluster) > 0 {
			return fmt.Errorf("hostgroup %s: contains both hostgroup and cluster", group.ID)
		}
		for _, group := range group.Sub {
			if err := hs.addGroup(group); err != nil {
				return err
			}
		}
	case len(group.Cluster) > 0:
		for _, cluster := range group.Cluster {
			if err := hs.addCluster(cluster); err != nil {
				return err
			}
		}
	}
	return nil
}

func (hs *Hosts) addCluster(cluster *Cluster) error {
	if err := cluster.Validate(); err != nil {
		return err
	}
	if err := hs.NS.Add(cluster.ID); err != nil {
		return err
	}
	hs.tags.Push(cluster.ID)
	defer hs.tags.Pop()
	switch {
	default:
		// ok
	case len(cluster.Host) > 0:
		for _, host := range cluster.Host {
			if err := hs.add(host, cluster); err != nil {
				return err
			}
		}
	}
	return nil
}

func (hs *Hosts) add(host *Host, cluster *Cluster) error {
	if err := host.Validate(); err != nil {
		return err
	}
	if err := hs.NS.Add(host.ID); err != nil {
		return err
	}
	hs.tags.Push(host.ID)
	defer hs.tags.Pop()
	host.ClusterID = cluster.ID
	host.Tags = hs.tags.Copy()
	hs.All = append(hs.All, host)
	return nil
}

// Filter is a system command that acts as a hook, allowing operators to serve
// custom rewrite/block filter rules.
type Filter struct {
	Path string `xml:"path,attr"`
}

// Run decodes into v json-encoded rules generated by the filter program.
// The filter receives program, host, and cluster details as command-line
// arguments.
func (f *Filter) Run(v interface{}, program, host, cluster string) error {
	err := f.run(v, program, host, cluster)
	if err != nil {
		return fmt.Errorf("config: filter error: %v", err)
	}
	return nil
}

func (f *Filter) run(v interface{}, program, host, cluster string) error {
	args := []string{f.Path, program, host, cluster}
	cmd := exec.Command(args[0], args[1:]...)
	stdout := new(bytes.Buffer)
	cmd.Stdout = stdout
	stderr := new(bytes.Buffer)
	cmd.Stderr = stderr
	defer func() {
		if stderr.Len() > 0 {
			log.Printf("config: filter stderr: %s", stderr)
		}
	}()
	if err := cmd.Run(); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := json.NewDecoder(stdout).Decode(v); err != nil {
		return err
	}
	return nil
}

// Namespace is a set of identifiers.
type Namespace map[string]bool

// Add adds an identifier to the namespace. If the name is already taken,
// an error is returned.
func (ns Namespace) Add(id string) error {
	if ns[id] {
		return fmt.Errorf("identifier redeclared: %v", id)
	}
	ns[id] = true
	return nil
}

type tags []string

func (ts *tags) Push(id string) { *ts = append(*ts, id) }
func (ts *tags) Pop()           { *ts = (*ts)[:len(*ts)-1] }

func (ts tags) Copy() []string {
	cp := make([]string, len(ts))
	copy(cp, ts)
	return cp
}

// An Element is an extended configuration element. Such elements are stored in
// Config for processing by extension modules.
type Element struct {
	Name  string      `xml:"-"`
	Value interface{} `xml:"-"`
	Raw   []byte      `xml:"-"`
}

func (e *Element) String() string {
	return fmt.Sprintf("%v", e.Value)
}

/*
func (e *Element) copy() *Element {
	return &Element{
		e.Name,
		e.Value,
		e.Raw,
	}
}
*/

func (e *Element) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := enc.EncodeToken(token); err != nil {
			return err
		}
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	e.Name = start.Name.Local
	e.Raw = buf.Bytes()
	return nil
}
