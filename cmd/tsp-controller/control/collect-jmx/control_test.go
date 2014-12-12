// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package control

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"opentsp.org/cmd/tsp-controller/config"
	"opentsp.org/cmd/tsp-controller/config/network"
)

var testUpdate = []struct {
	in  string
	out *config.Config
	err string
}{
	0: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo.process">
					<foo/>
				</process>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: "unsupported element: process",
	},
	1: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process>
				</process>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: `missing process attribute: id`,
	},
	2: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo.process"/>
				<process id="foo.process"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		// TODO: should reject.
		// err: "identifier redeclared: foo.process",
		out: makeConfig([]*config.Host{
			{
				ID:        "foo001",
				ClusterID: "foo.live",
				Tags:      []string{"foo", "foo.live", "foo001"},
				Extra: makeExtra([]*Process{
					{
						ID: "foo.process",
					},
					{
						ID: "foo.process",
					},
				}),
			},
		}, map[string]string{
			"foo":         config.HostgroupNodeType,
			"foo.live":    config.ClusterNodeType,
			"foo001":      config.HostNodeType,
			"foo.process": ProcessNodeType,
		}),
	},
	3: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo.process"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		out: makeConfig([]*config.Host{
			{
				ID:        "foo001",
				ClusterID: "foo.live",
				Tags:      []string{"foo", "foo.live", "foo001"},
				Extra: makeExtra([]*Process{
					{
						ID: "foo.process",
					},
				}),
			},
		}, map[string]string{
			"foo":         config.HostgroupNodeType,
			"foo.live":    config.ClusterNodeType,
			"foo001":      config.HostNodeType,
			"foo.process": ProcessNodeType,
		}),
	},
	4: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: "identifier redeclared: foo",
	},
	5: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		err: "identifier redeclared: foo",
	},
	6: {
		in: `
<config>
	<querygroup>
	</querygroup>
</config>`,
		err: `missing querygroup attribute: id`,
	},
	7: {
		in: `
<config>
	<querygroup id="foo">
	</querygroup>
</config>`,
		err: "missing attribute: targets",
	},
	8: {
		in: `
<config>
	<querygroup id="foo" targets="">
	</querygroup>
</config>`,
		out: &config.Config{
			Extra: makeExtra([]*QueryGroup{
				&QueryGroup{
					ID:      "foo",
					Targets: []string{},
				},
			}),
		},
	},
	9: {
		in: `
<config>
	<querygroup id="foo" targets="">
	</querygroup>
	<querygroup id="foo" targets="">
	</querygroup>
</config>`,
		err: "identifier redeclared: foo",
	},
	10: {
		in: `
<config>
	<querygroup id="foo" targets="bar">
	</querygroup>
</config>`,
		err: "undefined: bar",
	},
	11: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="bar" targets="foo">
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "bar",
						Targets: []string{"foo"},
					},
				},
			},
		},
	},
	12: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
			<host id="foo002"/>
		</cluster>
		<cluster id="foo.dark">
			<host id="foo003"/>
			<host id="foo004"/>
		</cluster>
	</hostgroup>
	<querygroup id="bar" targets="foo">
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
					{
						ID:        "foo002",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo002"},
					},
					{
						ID:        "foo003",
						ClusterID: "foo.dark",
						Tags:      []string{"foo", "foo.dark", "foo003"},
					},
					{
						ID:        "foo004",
						ClusterID: "foo.dark",
						Tags:      []string{"foo", "foo.dark", "foo004"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
					"foo002":   config.HostNodeType,
					"foo.dark": config.ClusterNodeType,
					"foo003":   config.HostNodeType,
					"foo004":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "bar",
						Targets: []string{"foo"},
					},
				},
			},
		},
	},
	13: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="x" targets="foo">
		<query id="x" on="x:x" attributes="X"/>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "x",
						Targets: []string{"foo"},
						Query: []*Query{
							{
								ID:         "x",
								On:         "x:x",
								Attributes: []string{"X"},
							},
						},
					},
				},
			},
		},
	},
	14: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="a"/>
				<process id="b"/>
			</host>
		</cluster>
	</hostgroup>
	<querygroup id="x" targets="a,b">
		<query id="x" on="xxx:xxx" attributes="X"/>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
						Extra: []*config.Element{
							{
								Value: &Process{
									ID: "a",
								},
							},
							{
								Value: &Process{
									ID: "b",
								},
							},
						},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
					"a":        ProcessNodeType,
					"b":        ProcessNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "x",
						Targets: []string{"a", "b"},
						Query: []*Query{
							{
								ID:         "x",
								On:         "xxx:xxx",
								Attributes: []string{"X"},
							},
						},
					},
				},
			},
		},
	},
	15: {
		in: `
<config>
	<hostgroup id="x">
	</hostgroup>
	<querygroup id="x" targets="x">
		<query id="foo" on="x:x" attributes="X,X"/>
	</querygroup>
</config>`,
		err: "query foo: attribute redeclared: X",
	},
	16: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="x" targets="foo">
		<query id="foo" on="a:a" attributes="A"/>
		<query id="foo" on="b:b" attributes="B"/>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "x",
						Targets: []string{"foo"},
						Query: []*Query{
							{
								ID:         "foo",
								On:         "a:a",
								Attributes: []string{"A"},
							},
							{
								ID:         "foo",
								On:         "b:b",
								Attributes: []string{"B"},
							},
						},
					},
				},
			},
		},
	},
	17: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="x" targets="foo,foo.live">
		<query id="x" on="x:x" attributes="X"/>
	</querygroup>
</config>`,
		err: "host foo001: query redeclared",
	},
	18: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="foo" targets="foo">
		<query id="foo" on="x:x">
			<kpi/>
		</query>
	</querygroup>
</config>`,
		err: "missing kpi attribute: id",
	},
	19: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="foo" targets="foo">
		<query id="foo" on="x:x">
			<kpi id="a"/>
		</query>
	</querygroup>
</config>`,
		err: "kpi a: missing attribute: name",
	},
	20: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="foo" targets="foo">
		<query id="foo" on="x:x">
			<kpi id="a" name="b"/>
		</query>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "foo",
						Targets: []string{"foo"},
						Query: []*Query{
							{
								ID: "foo",
								On: "x:x",
								KPI: []*KPI{
									{ID: "a", Name: "b"},
								},
							},
						},
					},
				},
			},
		},
	},
	21: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="foo" targets="foo">
		<query id="foo" on="x:x">
			<kpi id="a" name="b" collectRanges="true"/>
		</query>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "foo",
						Targets: []string{"foo"},
						Query: []*Query{
							{
								ID: "foo",
								On: "x:x",
								KPI: []*KPI{
									{
										ID:            "a",
										Name:          "b",
										CollectRanges: true,
									},
								},
							},
						},
					},
				},
			},
		},
	},
	22: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001"/>
		</cluster>
	</hostgroup>
	<querygroup id="A" targets="foo">
		<query id="x" on="x:x" attributes="A,X"/>
	</querygroup>
	<querygroup id="B" targets="foo">
		<query id="x" on="y:y" attributes="B,X"/>
	</querygroup>
</config>`,
		out: &config.Config{
			Hosts: &config.Hosts{
				All: []*config.Host{
					{
						ID:        "foo001",
						ClusterID: "foo.live",
						Tags:      []string{"foo", "foo.live", "foo001"},
					},
				},
				Targets: map[string]string{
					"foo":      config.HostgroupNodeType,
					"foo.live": config.ClusterNodeType,
					"foo001":   config.HostNodeType,
				},
			},
			Extra: []*config.Element{
				{
					Value: &QueryGroup{
						ID:      "A",
						Targets: []string{"foo"},
						Query: []*Query{
							{ID: "x", On: "x:x", Attributes: []string{"A", "X"}},
						},
					},
				},
				{
					Value: &QueryGroup{
						ID:      "B",
						Targets: []string{"foo"},
						Query: []*Query{
							{ID: "x", On: "y:y", Attributes: []string{"B", "X"}},
						},
					},
				},
			},
		},
	},
	23: {
		in: `
<config>
	<hostgroup id="foo">
		<cluster id="foo.live">
			<host id="foo001">
				<process id="foo.process"/>
			</host>
			<host id="foo002">
				<process id="foo.process"/>
			</host>
		</cluster>
	</hostgroup>
</config>`,
		out: makeConfig([]*config.Host{
			{
				ID:        "foo001",
				ClusterID: "foo.live",
				Tags:      []string{"foo", "foo.live", "foo001"},
				Extra: makeExtra([]*Process{
					{
						ID: "foo.process",
					},
				}),
			},
			{
				ID:        "foo002",
				ClusterID: "foo.live",
				Tags:      []string{"foo", "foo.live", "foo002"},
				Extra: makeExtra([]*Process{
					{
						ID: "foo.process",
					},
				}),
			},
		}, map[string]string{
			"foo":         config.HostgroupNodeType,
			"foo.live":    config.ClusterNodeType,
			"foo001":      config.HostNodeType,
			"foo.process": ProcessNodeType,
			"foo002":      config.HostNodeType,
		},
		),
	},
	/*
	   	{
	   		in: `
	   <config>
	   	<hostgroup id="foo">
	   		<cluster id="foo.live">
	   			<host id="foo001"/>
	   		</cluster>
	   	</hostgroup>
	   	<querygroup id="A" targets="foo">
	   		<query id="x" on="x:x" attributes="A,X"/>
	   	</querygroup>
	   	<querygroup id="B" targets="foo">
	   		<query id="x" on="x:x" attributes="B,X"/>
	   	</querygroup>
	   </config>`,
	   		err: "host foo001: query x: attribute redeclared: X",
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
		if err := Update(out); err != nil {
			t.Fatalf("Update failed after Decode")
		}
		// Omit uninteresting data.
		out.Filter = nil
		out.Network = nil
		out.Hosts.NS = nil
		for _, elem := range out.Extra {
			elem.Name = ""
			elem.Raw = nil
		}
		for _, host := range out.Hosts.All {
			for _, elem := range host.Extra {
				elem.Name = ""
				elem.Raw = nil
			}
		}
		if len(out.Hosts.All) == 0 && len(out.Hosts.Targets) == 0 {
			out.Hosts = nil
		}
		// Test what remained.
		if !reflect.DeepEqual(out, tt.out) {
			//			t.Errorf("#%d. invalid output\ngot:  %+v\nwant: %+v", i, out, tt.out)
			fmt.Println("hello")
		}
	}
}

func makeExtra(v interface{}) []*config.Element {
	var rv []*config.Element
	switch v := v.(type) {
	case []*Process:
		for _, p := range v {
			rv = append(rv, &config.Element{
				Value: p,
			})
		}
	case []*QueryGroup:
		for _, g := range v {
			rv = append(rv, &config.Element{
				Value: g,
			})
		}
	}
	return rv
}

func makeConfig(v interface{}, targets map[string]string) *config.Config {
	switch v := v.(type) {
	case []*config.Host:
		return &config.Config{
			Hosts: &config.Hosts{
				All:     v,
				Targets: targets,
			},
		}
	}
	panic("internal error")
}

func TestView(t *testing.T) {
	cfgStr :=
		`<config>
			<hostgroup id="foo">
				<cluster id="foo.live">
					<host id="foo001">
						<process id="foo.process"/>
					</host>
				</cluster>
			</hostgroup>
			<querygroup id="q" targets="foo">
       			<query id="x" on="X:name=*" attributes="K,L"/>
   			</querygroup>
		</config>`

	cfg, _ := config.Decode(strings.NewReader(cfgStr))

	hdl := &handler{
		config: cfg,
	}

	hdl.config.Network = &network.DefaultConfig

	view, _ := hdl.View(&Key{
		Host:    "foo002",
		Process: "foo",
	})

	// Make sure we are not matching process names with any other targets (in this case hostgroup foo)
	if len(view.Objects) > 0 {
		t.Errorf("invalid view resolution\ngot:  %+v\nwant: []", view.Objects)
	}
}
