// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package config implements shared configuration-related routines.
package config

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
	"time"
)

var Debug *log.Logger

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // TODO: make this configurable.
		},
	},
	Timeout: 15 * time.Second,
}

// getInterval determines frequency of server config lookups.
const getInterval = 60 * time.Second

// maxBytes limits the size of server config response.
const maxBytes = 1 << 20

var statLoadErrors = expvar.NewMap("config.load.Errors")

// Struct is an interface satisfied by every configuration struct. Reset resets
// struct by setting default values and zeroing others. Validate returns nil if
// the configuration is valid for processing by the client.
type Struct interface {
	Reset()
	Validate() error
}

var saved *loadPath

type loadPath struct {
	filePath   string
	serverPath string
}

// Load loads the config struct with values stored in the given file. The file's
// encoding is JSON.
//
// In addition, depending on environment details, a second file is loaded over
// network. The remote file overrides those settings of the local file which have
// been marked as dynamic in the corresponding struct field. To mark field as
// dynamic, tag it with config=dynamic, for example:
//
//	RelayHost string `config:"dynamic"`
//
// The server path must be a valid HTTP path except the order of query parameters
// is significant. The path may include the template action {{.Hostname}} which
// expands to local host name, for example:
//
//	"/tsp/forwarder?host={{.Hostname}}"
//
// The test for config=dynamic tag is not performed recursively.
//
// At least one path must be provided. If a path is empty, it is ignored.
func Load(config Struct, filePath, serverPath string) {
	if saved != nil {
		log.Panicf("Load called twice")
	}
	if serverPath != "" && os.Getenv("CONTROL_HOST") == "" {
		serverPath = ""
	}
	saved = &loadPath{filePath, serverPath}
	load(config, filePath, serverPath)
}

// Next blocks until new configuration values arrive over network. Once they do,
// the config struct is rewritten to contain the new settings.
//
// If called before Load, Next will panic. In some environments, Next may block
// forever.
func Next(config Struct) {
	if saved.serverPath == "" {
		select {}
	}
	load(config, saved.filePath, saved.serverPath)
}

func load(config Struct, filePath, serverPath string) {
	rateLimit := time.NewTicker(1 * time.Second)
	defer rateLimit.Stop()
	switch {
	default:
		log.Fatalf("config: no path defined")
	case filePath != "" && serverPath == "":
		config.Reset()
		ok := decodeFileTry(config, filePath)
		if !ok {
			os.Exit(1)
		}
		if !isValid(config) {
			os.Exit(1)
		}
	case filePath == "" && serverPath != "":
		client := dial(serverPath)
		for ; ; <-rateLimit.C {
			config.Reset()
			ok := decodeServerTry(config, client)
			if !ok {
				continue
			}
			if !isValid(config) {
				continue
			}
			break
		}
		defaultClient = client
	case filePath != "" && serverPath != "":
		client := dial(serverPath)
		for ; ; <-rateLimit.C {
			config.Reset()
			ok := decodeFileTry(config, filePath)
			if !ok {
				continue
			}
			ok = decodeServerTry(config, client)
			if !ok {
				continue
			}
			if !isValid(config) {
				continue
			}
			break
		}
		defaultClient = client
	}
}

func decodeFileTry(config Struct, path string) (ok bool) {
	if err := decodeFile(config, path); err != nil {
		statLoadErrors.Add("type=Decode", 1)
		log.Printf("config: file decode error: %v", err)
		return
	}
	ok = true
	return
}

func decodeFile(config Struct, path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if err := unmarshal(config, buf); err != nil {
		return err
	}
	return nil
}

func decodeServerTry(config Struct, client *client) (ok bool) {
	if err := decodeServer(config, client); err != nil {
		statLoadErrors.Add("type=Decode", 1)
		log.Printf("config: server decode error: %v", err)
		return
	}
	ok = true
	return
}

func decodeServer(config Struct, client *client) error {
	buf := client.NextUpdate()
	filteredStruct := newFieldFilter(config)
	if err := unmarshal(filteredStruct, buf); err != nil {
		return err
	}
	return nil
}

func unmarshal(config interface{}, buf []byte) error {
	if err := json.Unmarshal(buf, config); err != nil {
		err = addLineNum(err, buf)
		return err
	}
	return nil
}

func isValid(config Struct) bool {
	if err := config.Validate(); err != nil {
		statLoadErrors.Add("type=Validation", 1)
		log.Printf("config: validation error: %v", err)
		return false
	}
	return true
}

func addLineNum(err error, buf []byte) error {
	syntax, ok := err.(*json.SyntaxError)
	if !ok {
		return err
	}
	start := bytes.LastIndex(buf[:syntax.Offset], []byte("\n")) + 1
	num := bytes.Count(buf[:start], []byte("\n")) + 1
	return fmt.Errorf("line %d: %v", num, err)
}

var defaultClient *client

type client struct {
	addr, path string
	nextUpdate chan []byte
}

func dial(path string) *client {
	if defaultClient != nil {
		return defaultClient
	}
	c := &client{
		addr:       os.Getenv("CONTROL_HOST"),
		path:       path,
		nextUpdate: make(chan []byte),
	}
	go c.mainloop()
	return c
}

// NexUpdate blocks until a payload arrives that differs from the previous one.
func (c *client) NextUpdate() []byte {
	return <-c.nextUpdate
}

func (c *client) mainloop() {
	var last []byte
	tick := time.Tick(getInterval)
	for ; ; <-tick {
		buf, ok := c.getTry()
		if !ok {
			continue
		}
		if last != nil && bytes.Equal(buf, last) {
			continue
		}
		c.nextUpdate <- buf
		last = buf
	}
}

func (c *client) get() (buf []byte, err error) {
	resp, err := c.getPlainOrTLS()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := io.LimitReader(resp.Body, maxBytes)
	return ioutil.ReadAll(r)
}

func (c *client) requestURL(scheme string) string {
	i := strings.Index(c.path, "?")
	URL := url.URL{
		Scheme:   scheme,
		Host:     c.addr,
		Path:     path.Join("/control/v1", c.path[:i]),
		RawQuery: expand(c.path[i+1:]),
	}
	return URL.String()
}

func (c *client) getPlainOrTLS() (*http.Response, error) {
	requests := []string{
		c.requestURL("https"),
		c.requestURL("http"),
	}
	var (
		ok     *http.Response
		errors []error
	)
	for _, URL := range requests {
		if Debug != nil {
			Debug.Printf("get %s", URL)
		}
		resp, err := httpClient.Get(URL)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("Get %s: got status code %d (%q)",
				URL, resp.StatusCode, resp.Status)
			errors = append(errors, err)
			continue
		}
		ok = resp
		break
	}
	if ok == nil {
		return nil, fmt.Errorf("%q", errors)
	}
	return ok, nil
}

func expand(s string) string {
	var data struct {
		Hostname string
	}
	var err error
	data.Hostname, err = os.Hostname()
	if err != nil {
		log.Panic(err)
	}
	t := template.Must(template.New("query").Parse(s))
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		log.Panic(err)
	}
	return buf.String()
}

func (c *client) getTry() (buf []byte, ok bool) {
	buf, err := c.get()
	if err != nil {
		statLoadErrors.Add("type=Client", 1)
		log.Printf("config: client error: %v", err)
		return
	}
	ok = true
	return
}

// fieldFilter blocks illegal writes to Struct.
type fieldFilter struct {
	v     interface{}
	field []string
}

func newFieldFilter(v interface{}) *fieldFilter {
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	if typ.Kind() != reflect.Struct {
		panic("not a struct")
	}
	filter := &fieldFilter{v: v}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Tag.Get("config") == "dynamic" {
			filter.field = append(filter.field, f.Name)
		}
	}
	return filter
}

func (filter *fieldFilter) UnmarshalJSON(buf []byte) error {
	got := make(map[string]json.RawMessage)
	if err := json.Unmarshal(buf, &got); err != nil {
		return err
	}
	for _, f := range filter.field {
		raw := got[f]
		if raw == nil {
			continue
		}
		buf := []byte(fmt.Sprintf("{%q: %s}", f, raw))
		if err := json.Unmarshal(buf, filter.v); err != nil {
			return err
		}
	}
	return nil
}
