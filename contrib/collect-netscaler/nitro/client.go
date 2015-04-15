// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package nitro

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

// maxTrips limits the number of data round trips to the Nitro API on behalf of
// single call. In the extreme case, a total of 2*maxTrips will be made: two
// data round trips, and two corresponding cookie refresh trips.
const maxTrips = 2

// responseTimeout limits the wait time for response of single HTTP round trip.
const responseTimeout = 5 * time.Second

// MaxConnsPerHost limits the number of connections simultaneously established
// by Client. The limit exists to help avoid nitro error 446 (Connection limit
// to CFE exceeded).
const MaxConnsPerHost = 16

// Nitro-level status codes
const (
	statusOK    = 0
	statusAuth1 = 354  // Invalid username or password
	statusAuth2 = 444  // Session expired or killed. Please login again
	statusAuth3 = 2138 // Not authorized to execute this command
)

// Statistics
var (
	statClientRequests  = expvar.NewMap("nitro.client.Requests")
	statClientResponses = expvar.NewMap("nitro.client.Responses")
	statClientErrors    = expvar.NewMap("nitro.client.Errors")
	statClientMillis    = expvar.NewMap("nitro.client.Millis")
)

// Client wraps http.Client in order to provide access to the Nitro API.
type Client struct {
	client *http.Client
	conn   connLimit
	addr   string
	auth   struct {
		username string
		password string
		cookie   cookie
	}

	Config ConfigService
	Stat   StatService
}

// NewClient returns Nitro API client that will request all resources from addr
// using the provided HTTP client.
//
// For maximum efficiency, pass custom http.Client with http.Transport's
// MaxIdleConnsPerHost set to MaxConnsPerHost. If nil client is passed, such
// optimal client will be allocated automatically.
//
// The client remains valid for use despite any errors encountered.
func NewClient(client *http.Client, addr, username, password string) *Client {
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: MaxConnsPerHost,
			},
		}
	}
	c := &Client{
		client: client,
		conn:   make(connLimit, MaxConnsPerHost),
		addr:   addr,
	}
	c.Stat.client = c
	c.Config.client = c
	c.auth.username = username
	c.auth.password = password
	return c
}

// Close releases all resources held by the Client.
func (c *Client) Close() {
	c.client.Transport.(*http.Transport).CloseIdleConnections()
}

func (c *Client) url() string {
	return fmt.Sprintf("https://%s/nitro/v1", c.addr)
}

func (c *Client) do(r response, path string) (err error) {
	statKey := fmt.Sprintf("addr=%s path=%s", c.addr, cleanPath(path))
	start := time.Now()
	defer func() {
		statKey += fmt.Sprintf(" error=%v", err != nil)
		statClientResponses.Add(statKey, 1)
		statClientMillis.Add(statKey, time.Since(start).Nanoseconds()/1e6)
	}()

	var errors []error

	for i := 0; i < maxTrips; i++ {
		if err := c.auth.cookie.Refresh(c); err != nil {
			errors = append(errors, err)
			break
		}
		req := newRequest(c.url(), path, c.auth.cookie.Get())
		if err := c.roundtrip(r, req); err != nil {
			errors = append(errors, err)
			if _, ok := err.(authError); ok {
				log.Printf("nitro: attempting cookie refresh due to authentication error: %q", err)
				c.auth.cookie.Reset()
				continue
			}
			break
		}
		errors = nil
		break
	}

	if len(errors) > 0 {
		err := fmt.Errorf("request error: %v", join(errors))
		return err
	}

	return nil
}

func (c *Client) roundtrip(r response, req *http.Request) error {
	c.conn.Wait()
	defer c.conn.Done()

	respChan := make(chan httpResponse, 1)

	go func() {
		resp, err := c.client.Do(req)
		respChan <- httpResponse{resp, err}
	}()

	var resp httpResponse
	select {
	case <-time.After(responseTimeout):
		c.client.Transport.(*http.Transport).CancelRequest(req)
		statClientErrors.Add("type=httpTimeout", 1)
		return fmt.Errorf("Get %s: response timeout", req.URL)

	case resp = <-respChan:
		// ok
	}

	if resp.error != nil {
		statClientErrors.Add("type=httpTransport", 1)
		return resp.error
	}

	defer func() {
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()

	switch code := resp.StatusCode; code {
	default:
		err := fmt.Errorf("got http code %d (%s)", code, http.StatusText(code))
		statClientErrors.Add("type=httpStatus", 1)
		return err

	case http.StatusUnauthorized:
		err := fmt.Errorf("got http code %d (%s)", code, http.StatusText(code))
		statClientErrors.Add("type=httpStatus", 1)
		return authError{err}

	case http.StatusOK:
		// ok

	case http.StatusCreated:
		// ok
	}

	reader := errorChecker{Reader: resp.Body}
	if err := json.NewDecoder(&reader).Decode(r); err != nil {
		if reader.Err {
			statClientErrors.Add("type=httpTransport", 1)
		} else {
			statClientErrors.Add("type=JSON", 1)
		}
		return err
	}

	switch code := r.errorCode(); code {
	default:
		err := fmt.Errorf("Get %s: server error: %s (code %d)",
			resp.Request.URL, r.message(), code)
		statClientErrors.Add("type=Server", 1)
		return err

	case statusAuth1, statusAuth2, statusAuth3:
		err := fmt.Errorf("Get %s: server auth error: %s (code %d)",
			resp.Request.URL, r.message(), code)
		statClientErrors.Add("type=ServerAuth", 1)
		return authError{err}

	case statusOK:
		// ok
	}

	return nil
}

type cookie struct {
	*http.Cookie
	mu sync.Mutex
}

func (c *cookie) Get() *http.Cookie {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Cookie
}

func (c *cookie) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Cookie = nil
}

func (c *cookie) Refresh(client *Client) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Cookie != nil {
		return nil
	}
	cookie, err := newCookie(client)
	if err != nil {
		return fmt.Errorf("cookie refresh error: %v", err)
	}
	c.Cookie = cookie
	log.Printf("nitro: cookie refresh hash=%s", c)
	return nil
}

// String returns a crypto hash of the cookie, which prevents leaking auth
// secret via the log file.
func (c *cookie) String() string {
	in := []byte(c.Cookie.String())
	out := sha1.Sum(in)
	return hex.EncodeToString(out[:])
}

func newCookie(c *Client) (*http.Cookie, error) {
	var r responseSessionID
	req := newSessionRequest(c.url(), c.auth.username, c.auth.password)
	if err := c.roundtrip(&r, req); err != nil {
		return nil, err
	}
	return &http.Cookie{Name: "sessionid", Value: r.SessionID}, nil
}

func newSessionRequest(url, username, password string) *http.Request {
	dumper := &requestDumper{}
	client := &http.Client{Transport: dumper}
	client.PostForm(url, newSessionForm(username, password))
	return dumper.req
}

func newSessionForm(username, password string) url.Values {
	var data struct {
		Login struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"login"`
	}
	data.Login.Username = username
	data.Login.Password = password
	buf, err := json.Marshal(data)
	if err != nil {
		log.Panicf("internal error: %v", err)
	}
	return url.Values{"object": []string{string(buf)}}
}

func newRequest(url, path string, cookie *http.Cookie) *http.Request {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", url, path), nil)
	if err != nil {
		log.Panicf("internal error: %v", err)
	}
	req.AddCookie(cookie)
	return req
}

// cleanPath ensures path p is safe for use as expvar tag value.
func cleanPath(p string) string {
	switch {
	default:
		return p

	case strings.HasPrefix(p, "config/"):
		return path.Dir(p)
	}
}

// connLimit is a semaphore used to limit the number of active connections.
type connLimit chan bool

func (cl connLimit) Wait() { cl <- true }
func (cl connLimit) Done() { <-cl }

type authError struct {
	error
}

type httpResponse struct {
	*http.Response
	error
}

// requestDumper is a http.RoundTripper that can be used to intercept
// requests created by http.Client.
type requestDumper struct {
	req *http.Request
}

func (rd *requestDumper) RoundTrip(req *http.Request) (*http.Response, error) {
	rd.req = req
	return nil, errors.New("dummy transport")
}

// join combines multiple errors.
func join(errors []error) error {
	var s []string
	for _, err := range errors {
		s = append(s, err.Error())
	}
	err := fmt.Errorf("%s", strings.Join(s, ": "))
	return err
}

// errorChecker sets Err to true if the underlying io.Reader encountered an
// error.
type errorChecker struct {
	io.Reader
	Err bool
}

func (r *errorChecker) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	if err != nil {
		r.Err = true
	}
	return n, err
}
