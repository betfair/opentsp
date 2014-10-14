// Copyright 2014 The Sporting Exchange Limited. All rights reserved.
// Use of this source code is governed by a free license that can be
// found in the LICENSE file.

package statse

import (
	"fmt"

	"opentsp.org/internal/github.com/vaughan0/go-zmq"
)

// ListenAddr is the listen address of Statse sink service.
const ListenAddr = "tcp://127.0.0.1:14444/"

type listener struct {
	socket *zmq.Socket
	ch     <-chan [][]byte
}

func Listen(addr string) (Reader, error) {
	socket, err := zmq.DefaultContext().Socket(zmq.Sub)
	if err != nil {
		return nil, fmt.Errorf("statse: error creating zmq socket: %v", err)
	}
	if err := socket.Bind(addr); err != nil {
		return nil, fmt.Errorf("statse: error binding zmq socket: %v", err)
	}
	socket.Subscribe([]byte(""))
	return &listener{socket, socket.Channels().In()}, nil
}

func (l *listener) ReadParts() [2][]byte {
	parts := <-l.ch
	return [2][]byte{
		parts[0],
		parts[1],
	}
}
