// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package control

// if no aggregator specified, statse forwarder should do local aggregation

import (
	"reflect"
	"strings"
	"testing"

	"opentsp.org/cmd/tsp-controller/config"
)

var testUpdate = []struct {
	in  string
	out *config.Config
	err string
}{
/*
	{
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<statse aggregator="true"/>
			</host>
			<host id="foo002">
				<statse aggregator="true"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: "cluster foo.live: statse: aggregator redeclared",
	},
*/
}

func TestUpdate(t *testing.T) {
	for i, tt := range testUpdate {
		out, err := config.Decode(strings.NewReader(tt.in))
		if err != nil {
			if tt.err == "" {
				t.Errorf("#%d. unexpected error: %v", i, err)
				continue
			}
			if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("#%d. invalid error, got: %s, want: %s", i, err, tt.err)
				continue
			}
			continue
		}
		if tt.err != "" {
			t.Errorf("#%d. unexpected success, want error: %v", i, tt.err)
			continue
		}
		out.Filter = nil
		out.Network = nil
		if len(out.Hosts.All) == 0 {
			out.Hosts = nil
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("#%d. invalid output\ngot:  %+v\nwant: %+v", i, out, tt.out)
		}
	}
}

func makeConfig(v interface{}) *config.Config {
	switch v := v.(type) {
	case []*config.Host:
		return &config.Config{
			Hosts: &config.Hosts{
				All: v,
			},
		}
	}
	panic("internal error")
}
