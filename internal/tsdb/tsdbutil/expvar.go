// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package tsdbutil

import (
	"encoding/json"
	"expvar"
	"log"
	"strings"
	"time"

	"opentsp.org/internal/tsdb"
)

func ExportVars(t time.Time, fn func(*tsdb.Point)) {
	expvar.Do(func(kv expvar.KeyValue) {
		metric := kv.Key
		switch v := kv.Value.(type) {
		default:
			// log.Printf("tsdbutil: ExportVars: ignoring unsupported type %T for %s", v, kv.Key)

		case expvar.Func:
			if kv.Key != "memstats" {
				return
			}
			data := make(map[string]interface{})
			dec := json.NewDecoder(strings.NewReader(v.String()))
			dec.UseNumber()
			if err := dec.Decode(&data); err != nil {
				log.Print(err)
				return
			}
			for key, value := range data {
				num, ok := value.(json.Number)
				if !ok {
					continue
				}
				n, err := num.Int64()
				if err != nil {
					log.Print(err)
					continue
				}
				point, err := tsdb.NewPoint(t, n, "mem."+key)
				if err != nil {
					log.Printf("tsdbutil: ExportVars: cannot export memstats.%s: %v", key, err)
					continue
				}
				fn(point)
			}

		case *expvar.Int:
			value := kv.Value.String()
			point, err := tsdb.NewPoint(t, value, metric)
			if err != nil {
				log.Panic(err)
			}
			fn(point)

		case *expvar.Float:
			value := kv.Value.String()
			point, err := tsdb.NewPoint(t, value, metric)
			if err != nil {
				log.Panic(err)
			}
			fn(point)

		case *expvar.Map:
			v.Do(func(kv expvar.KeyValue) {
				switch kv.Value.(type) {
				default:
					// log.Printf("tsdbutil: ExportVars: ignoring unsupported type %T for key %s of %s", v2, kv2.Key, kv.Key)

				case expvar.Func:
					tags := kv.Key
					value := kv.Value
					tags = strings.Replace(tags, "=", " ", -1)
					point, err := tsdb.NewPoint(t, 0, metric, strings.Fields(tags)...)
					if err != nil {
						log.Panic(err)
					}
					if err := point.SetValue(value.String()); err != nil {
						log.Panic(err)
					}
					fn(point)

				case *expvar.Int:
					tags := kv.Key
					value := kv.Value
					tags = strings.Replace(tags, "=", " ", -1)
					point, err := tsdb.NewPoint(t, 0, metric, strings.Fields(tags)...)
					if err != nil {
						log.Panic(err)
					}
					if err := point.SetValue(value.String()); err != nil {
						log.Panic(err)
					}
					fn(point)

				case *expvar.Float:
					tags := kv.Key
					value := kv.Value
					tags = strings.Replace(tags, "=", " ", -1)
					point, err := tsdb.NewPoint(t, 0, metric, strings.Fields(tags)...)
					if err != nil {
						log.Panic(err)
					}
					if err := point.SetValue(value.String()); err != nil {
						log.Panic(err)
					}
					fn(point)
				}
			})
		}
	})
}
