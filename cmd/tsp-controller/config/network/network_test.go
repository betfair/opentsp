// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package network

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

var testUnmarshal = []struct {
	in  string
	out Config
}{
	0: {
		in: `
<network>
</network>
`,
		out: Config{},
	},
	1: {
		in: `
<network>
	<restrict host="foo"/>
	<aggregator host="ahost"/>
	<poller host="phost"/>
	<subscriber id="s" host="shost" direct="true" dedup="true"/>
</network>
`,
		out: Config{
			Aggregator: &Aggregator{
				Host: "ahost",
			},
			Poller: &Poller{
				Host: "phost",
			},
			Subscriber: []*Subscriber{
				{
					ID:     "s",
					Host:   "shost",
					Direct: true,
					Dedup:  true,
				},
			},
		},
	},
}

func TestUnmarshal(t *testing.T) {
	for i, tt := range testUnmarshal {
		f, err := ioutil.TempFile("", "unmarshaltest")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())
		if _, err := f.Write([]byte(tt.in)); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		var config Config
		r := strings.NewReader(`<network path="` + f.Name() + `"/>`)
		if err := xml.NewDecoder(r).Decode(&config); err != nil {
			t.Fatal(err)
		}
		config.restrict = nil
		if !reflect.DeepEqual(config, tt.out) {
			t.Errorf("#%d. got %+v, want %+v", i, config, tt.out)
		}
	}
}

func TestFileAbsent(t *testing.T) {
	var config Config
	r := strings.NewReader(`<network path="/var/empty/foo"/>`)
	if err := xml.NewDecoder(r).Decode(&config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(config, DefaultConfig) {
		t.Errorf("got %v, want %v", config, DefaultConfig)
	}
}
