// Copyright 2018 The gRPC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package socketpair

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Listener implements a listener that creates local, buffered net.Pipe-backed
// net.Conns via its Accept and Dial methods.
type Listener struct {
	mu   sync.Mutex
	done chan struct{}
	conn chan net.Conn
}

// Listen returns a Listener that can only be contacted by its own Dialers and
// creates buffered connections between the two.
func Listen() *Listener {
	return &Listener{
		done: make(chan struct{}),
		conn: make(chan net.Conn),
	}
}

// Accept blocks until Dial or Close are called. If Dial is called, it returns
// a net.Conn for the server half of the connection.
func (l *Listener) Accept() (net.Conn, error) {
	select {
	case <-l.done:
		return nil, fmt.Errorf("use of closed network connection")
	case conn := <-l.conn:
		return conn, nil
	}
}

// Close stops the listener. Close unblocks Accept and Dial.
func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	select {
	case <-l.done:
		// Already closed.
		break
	default:
		close(l.done)
	}
	return nil
}

func (l *Listener) Addr() net.Addr { return nil }

// Dial creates an in-memory full-duplex network connection, unblocks Accept by
// providing it the server half of the connection, and returns the client half
// of the connection.
func (l *Listener) Dial() (net.Conn, error) {
	return l.DialContext(context.Background())
}

// DialContext creates an in-memory full-duplex network connection, unblocks
// Accept by providing it the server half of the connection, and returns the
// client half of the connection. If ctx is Done, returns ctx.Err()
func (l *Listener) DialContext(ctx context.Context) (net.Conn, error) {
	p1, p2 := net.Pipe()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-l.done:
		return nil, fmt.Errorf("closed")
	case l.conn <- p1:
		return p2, nil
	}
}
