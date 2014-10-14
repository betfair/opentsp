// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package config

import "net/http"

func init() {
	handler := handler{}
	http.Handle("/config/v1/validate", &handler)
	http.Handle("/config/validate", &handler) // legacy
}

var validateJobLimit = make(validateSem, 10)

// handler implements configuration validation service.
type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	default:
		http.NotFound(w, req)
	case "/config/v1/validate", "/config/validate":
		h.Validate(w, req)
	}
}

// Validate tests the incoming configuration file.
func (h *handler) Validate(w http.ResponseWriter, req *http.Request) {
	if m := req.Method; m != "PUT" {
		err := validateError("invalid method: " + m)
		http.Error(w, err, http.StatusBadRequest)
		return
	}
	validateJobLimit.Wait()
	defer validateJobLimit.Done()
	w.Header().Set("Content-Type", "text/plain")
	_, err := Decode(req.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("ok"))
	}
}

func validateError(error string) string {
	return error +
		"\n" +
		"usage: curl -qsS --max-time 10 --request PUT --data-binary @- http://host:8084/config/v1/validate < file"
}

type validateSem chan bool

func (ch validateSem) Wait() { ch <- true }
func (ch validateSem) Done() { <-ch }
