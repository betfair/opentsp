// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package control provides controller extension modules
package control

import (
	"encoding/json"
	"net/http"

	"opentsp.org/cmd/tsp-controller/config"
)

type module struct {
	update, handle func(*config.Config) error
}

var modules []*module

// Register is used by modules to register themselves.
//
// The update function updates the config to include module-specific
// settings. It typically decodes ExtraRaw fields, storing decoded data in
// the corresponding Extra field.
//
// The handle function registers module-specific HTTP handlers. It is
// passed a config that has already been updated using the update
// function.
func Register(update, handle func(*config.Config) error) {
	modules = append(modules, &module{update, handle})
	config.Register(update)
}

// ListenAndServe is a wrapper for http.ListenAndServe that starts
// handlers for all registered control modules prior to starting the
// server. The server caches succesful responses forever.
func ListenAndServe(addr string, config *config.Config) error {
	for _, m := range modules {
		if err := m.handle(config); err != nil {
			return err
		}
	}
	return http.ListenAndServe(addr, nil)
}

// Marshal is used by modules to marshal response to a view request.
func Marshal(w http.ResponseWriter, view interface{}) {
	w.Header().Set("Content-Type", "application/json")
	buf, err := json.Marshal(view)
	if err != nil {
		panic(err)
	}
	w.Write(buf)
}
