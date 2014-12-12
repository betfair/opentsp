// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package config

import (
	"reflect"
	"strings"
	"testing"
	"fmt"
)

var testDecode = []struct {
	in  string
	out *Config
	err string
}{
	0: {
		in: `
<config>
</config>`,
		out: &Config{},
	},
	1: {
		in: `
<config>
	<foo/>
</config>`,
		err: "unsupported element: foo",
	},
	2: {
		in: `
<config>
	<host id="foo"/>
</config>`,
		err: "unsupported element: host",
	},
	3: {
		in: `
<config>
	<cluster id="foo">
	</cluster>
</config>`,
		err: "unsupported element: cluster",
	},
	4: {
		in: `
<config>
	<hostgroup>
	</hostgroup>
</config>`,
		err: "missing hostgroup attribute: id",
	},
	5: {
		in: `
<config>
	<hostgroup id="foo">
	</hostgroup>
</config>`,
		out:makeConfig(([]*Host)(nil),
			map[string]string{
				"foo": HostgroupNodeType,
			}),
	},
	6: {
		in: `
<config>
	<hostgroup id="foo">
	</hostgroup>
	<hostgroup id="foo">
	</hostgroup>
</config>`,
		err: "identifier redeclared: foo",
	},
	7: {
		in: `
<config>
	<hostgroup id="foo">
		<host id="foo001"/>
	</hostgroup>
</config>`,
		err: "invalid element: host",
	},
	8: {
		in: `
<config>
	<hostgroup id="foo">
		<foo/>
	</hostgroup>
</config>`,
		err: "invalid element: foo",
	},
	9: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<foo/>
		</cluster>
	</hostgroup>
</config>`,
		err: "invalid element: foo",
	},
	10: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster>
		</cluster>
	</hostgroup>
</config>`,
		err: `missing cluster attribute: id`,
	},
	11: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
		</cluster>
	</hostgroup>
</config>`,
		out: makeConfig(([]*Host)(nil),
			map[string]string{
				"foo": HostgroupNodeType,
				"foo.live": ClusterNodeType,
			}),
	},
	12: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
		</cluster>
		<cluster id="foo.live">
		</cluster>
	</hostgroup>
</config>`,
		err: "identifier redeclared: foo.live",
	},
	13: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<foo/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: "unsupported element: foo",
	},
	14: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host/>
		</cluster>
	</hostgroup>
</config>`,
		err: `missing host attribute: id`,
	},
	15: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
			<host id="foo001"/>
		</cluster>
	</hostgroup>
</config>`,
		err: "identifier redeclared: foo001",
	},
	16: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
</config>`,
		out: makeConfig([]*Host{
			{
				ID:        "foo001",
				ClusterID: "foo.live",
				Tags:      []string{"foo", "foo.live", "foo001"},
			},
		}, map[string]string{
			"foo": HostgroupNodeType,
			"foo.live": ClusterNodeType,
			"foo001": HostNodeType,
		}),
	},
}

func makeConfig(v interface{}, targets map[string]string) *Config {
	switch v := v.(type) {
	case []*Host:
		return &Config{
			Hosts: &Hosts{
				All: v,
				Targets: targets,
			},
		}
	default: panic(fmt.Sprintf("unrecognized type %v", reflect.TypeOf(v)))
	}
}

func TestDecode(t *testing.T) {
	for i, tt := range testDecode {
		dec := newDecoder(strings.NewReader(tt.in))
		out, err := dec.Decode()
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
		out.Hosts.NS = nil
		out.Hosts.tags = nil
		if len(out.Hosts.All) == 0 && len(out.Hosts.Targets) == 0 {
			out.Hosts = nil
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("#%d. invalid output\ngot:  %+v\nwant: %+v", i, out, tt.out)
		}
	}
}
