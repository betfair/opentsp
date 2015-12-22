// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package config

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"opentsp.org/contrib/collect-netscaler/collect"
	"opentsp.org/contrib/collect-netscaler/nitro"
	"opentsp.org/internal/version"
)

var (
	aflag       = flag.String("a", "/dev/stdin", "username and password, json-encoded file")
	Verbose     = flag.Bool("v", false, "verbose mode")
	VersionMode = flag.Bool("version", false, "echo version and exit")
	Interval    = flag.Duration("i", 7*time.Second, "poll interval")
	unsafe      = flag.Bool("u", false, "disable tls")
)

var addr string // host[:port]

func Host() string {
	host := addr
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host
}

func init() {
	flag.Parse()
	if *VersionMode {
		fmt.Println(version.String())
		os.Exit(0)
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	addr = flag.Arg(0)
	username, password := decode()

	log.SetOutput(os.Stderr)

	nitro.Verbose = *Verbose

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: nitro.MaxConnsPerHost,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: *unsafe,
			},
		},
	}
	collect.Client = nitro.NewClient(client, addr, username, password)
}

func decode() (username, password string) {
	buf, err := ioutil.ReadFile(*aflag)
	if err != nil {
		log.Fatalf("error accessing auth file: %v", err)
	}
	var auth struct {
		Username string
		Password string
	}
	if err := json.Unmarshal(buf, &auth); err != nil {
		log.Fatalf("error parsing auth file: %v", err)
	}
	if auth.Username == "" || auth.Password == "" {
		log.Fatalf("invalid auth file: missing Username or Password")
	}
	return auth.Username, auth.Password
}
