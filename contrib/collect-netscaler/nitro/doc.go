// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

// Package nitro implements access to Citrix NetScaler NITRO API.
//
// Example:
//
// 	client := nitro.NewClient(nil, host, username, password)
// 	resp, err := client.Stat.Get("protocoltcp")
// 	if err != nil {
// 		...
// 	}
// 	fmt.Println("active tcp server connections:", resp.ProtocolTCP.ActiveServerConn)
package nitro

// Verbose controls verbose mode.
var Verbose = false
