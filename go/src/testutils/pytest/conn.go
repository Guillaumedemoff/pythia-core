// Copyright 2013 The Pythia Authors.
// This file is part of Pythia.
//
// Pythia is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// Pythia is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Pythia.  If not, see <http://www.gnu.org/licenses/>.

package pytest

import (
	"net"
	"pythia"
	"reflect"
	"testing"
	"time"
)

// Conn is a wrapper around pythia.Conn to log connection messages and check
// received messages.
type Conn struct {
	T    *testing.T
	Conn *pythia.Conn
}

// Dial establishes a test connection.
func Dial(t *testing.T, addr net.Addr) (*Conn, error) {
	conn, err := pythia.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &Conn{t, conn}, nil
}

// DialRetry establishes a test connection, retrying indefinitely.
func DialRetry(t *testing.T, addr net.Addr) *Conn {
	return &Conn{t, pythia.DialRetry(addr)}
}

// Send sends a message through the connection.
func (c *Conn) Send(msg pythia.Message) {
	c.T.Log(">>", msg)
	c.Conn.Send(msg)
}

// Expect reads len(expected) messages from the connection.
// Fail if a message was not in the expected set (order does not matter) or an
// expected message was not received.
// If nothing is received for more than timeout seconds, fail and continue.
func (c *Conn) Expect(timeout int, expected ...pythia.Message) {
	ok := make([]bool, len(expected))
	for i := 0; i < len(expected); i++ {
		select {
		case msg := <-c.Conn.Receive():
			found := false
			for i, m := range expected {
				if !ok[i] && reflect.DeepEqual(msg, m) {
					c.T.Log("<<", msg)
					ok[i] = true
					found = true
					break
				}
			}
			if !found {
				c.T.Error("<<(unexpected)", msg)
			}
		case <-time.After(time.Duration(timeout) * time.Second):
			c.T.Error("<< timed out")
			break
		}
	}
	for i, msg := range expected {
		if !ok[i] {
			c.T.Error("<<(missing)", msg)
		}
	}
}

// Close checks for an unexpected message waiting in the input channel and
// closes the connection.
func (c *Conn) Close() {
	select {
	case msg := <-c.Conn.Receive():
		c.T.Error("<<(unexpected)", msg)
	default:
	}
	c.Conn.Close()
}

// vim:set sw=4 ts=4 noet:
