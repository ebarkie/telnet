// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package telnet

import "log/slog"

// readState is the state of the Reader.
type readState uint

const (
	rsData   readState = iota // Reading data
	rsIAC                     // Interpret as Command received
	rsInd                     // Will, won't, do, or don't indicator received
	rsSub                     // Subnegotiation
	rsSubIAC                  // Interpret as Command within subnegotiation
)

// Read is a telnet Reader.
//
// The Reader provided by the client is read and command sequences
// are parsed and executed.  If a non-empty buffer is passed then
// data will be passed as-is.
func (t *Ctx) Read(b []byte) (n int, err error) {
	if len(b) > 0 {
		// Drain pending data from a previous empty-buffer Read first.
		if len(t.pending) > 0 {
			n = copy(b, t.pending)
			t.pending = t.pending[n:]
			return
		}

		// The user is expecting at least one byte to be returned so keep
		// reading until we get some data.
		for err == nil && n < 1 {
			n, err = t.read(b)
		}
	} else {
		buf := make([]byte, 16)
		n, err = t.read(buf)
		if n > 0 {
			t.pending = append(t.pending, buf[:n]...)
			n = 0
		}
	}

	return
}

func (t *Ctx) read(b []byte) (n int, err error) {
	buf := make([]byte, len(b))
	var num int
	num, err = t.rw.Read(buf)

	t.mu.Lock()
	for i := 0; i < num; i++ {
		switch t.rs {
		case rsData:
			// In data mode if an Interpret as Command is not received then
			// pass through the data to the read buffer.
			switch buf[i] {
			case byte(iac):
				// Begin IAC
				t.rs = rsIAC
			default:
				// Data
				b[n] = buf[i]
				n++
			}
		case rsIAC:
			// An Interpret as Command was received.
			cmd := Command(buf[i])
			switch cmd {
			case AYT:
				t.rw.Write([]byte("I am here"))
				t.rs = rsData
			case iac:
				// Escaped IAC
				b[n] = buf[i]
				n++

				t.rs = rsData
			case will, wont, do, dont:
				t.cb = []byte{buf[i]}
				t.rs = rsInd
			case sb:
				t.cb = []byte{}
				t.rs = rsSub
			case NOP:
				// No operation
			default:
				slog.Debug("ignoring unhandled command", "cmd", cmd)
				t.rs = rsData
			}
		case rsInd:
			// A will, won't, do, or don't indicated was previously given so
			// negotiate the option.
			t.negotiate(Command(t.cb[0]), buf[i])

			t.rs = rsData
		case rsSub:
			// Subnegotiation is in progress so keep accepting data until an
			// Interpret as Command is received.
			switch buf[i] {
			case byte(iac):
				t.rs = rsSubIAC
			default:
				t.cb = append(t.cb, buf[i])
			}
		case rsSubIAC:
			// An Interpret as Command received during subnegotiation is only
			// valid to receive another (escaped) IAC or a subnegotiation
			// end.
			switch Command(buf[i]) {
			case iac:
				// Escaped IAC
				t.cb = append(t.cb, buf[i])
			case se:
				t.subnegotiate(t.cb[0], t.cb[1:])

				t.rs = rsData
			default:
				slog.Error("unexpected byte in subnegotiation", "byte", buf[i])
				t.rs = rsIAC
			}
		}
	}
	t.mu.Unlock()

	return
}

// Write is a telnet Writer.
//
// Any Interpret as Command bytes are escaped and the result is written
// using the Writer provided by the client.
func (t *Ctx) Write(b []byte) (int, error) {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		buf = append(buf, c)
		if c == byte(iac) {
			buf = append(buf, c)
		}
	}

	t.mu.Lock()
	_, err := t.rw.Write(buf)
	t.mu.Unlock()

	if err != nil {
		return 0, err
	}
	return len(b), nil
}
