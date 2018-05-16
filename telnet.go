// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package telnet implements the RFC854 Telnet Protocol
// Specification, as well as:
//
//  RFC855  Telnet Option Specifications
//  RFC1143 The Q Method of Implementing TELNET Option Negotiation
package telnet

import "io"

//go:generate stringer -type Command

// Command is a telnet command.
type Command byte

// RFC854 telnet commands.
const (
	EOF  Command = 236 + iota // End of File character
	SP                        // Suspend Process
	AP                        // Abort Process
	EOR                       // End Of Record
	se                        // Subnegotiation End
	NOP                       // No Operation
	DM                        // Data Mark
	BRK                       // Break
	IP                        // Interrupt Process
	AO                        // Abort Output
	AYT                       // Are You There
	EC                        // Erase Character
	EL                        // Erase Line
	GA                        // Go Ahead
	sb                        // Subnegotiation Begin
	will                      // Desire to begin performing or confirmation nsNow performing
	wont                      // Refusal to perform or continue to perform
	do                        // Request other party perform or confirm expecting
	dont                      // Demand other party stop or confirm nsNo longer expecting
	iac                       // Interpret as Command
)

// Option is an interface for implementing telnet options.
type Option interface {
	// Byte returns the byte code of the option.
	Byte() byte
	// String is the text name of the option (for debugging).
	String() string

	// LetHim indicates if he is allowed to enable the option.
	LetHim() bool
	// LetUs indicates if we are willing to enable the option.
	LetUs() bool

	// Params is called when subnegotiation parameters are received.
	Params(tn *Ctx, params []byte)

	// SetHim sets the option state for him.
	SetHim(tn *Ctx, enabled bool)
	// SetUs sets the option state for us.
	SetUs(tn *Ctx, enabled bool)
}

// optState is the state of an option.
type optState struct {
	opt     Option
	him, us negState
}

// optStates holds the state of each option.
type optStates map[byte]optState

// load retrieves the state of an option code.
func (os optStates) load(code byte) optState {
	s, found := os[code]
	if !found {
		s.opt = noOpt{Code: code}
	}

	return s
}

// store updates the state of an option.
func (os optStates) store(s optState) {
	os[s.opt.Byte()] = s
}

// Ctx is a telnet context.
type Ctx struct {
	// rw is the ReadWriter provided by the client.  Since telnet runs
	// over TCP this is typically a net.Conn.
	rw io.ReadWriter

	// rs is the Reader state.  It's either reading data or in various stages
	// of parsing a command.
	rs readState
	// cb is the Reader command buffer.  It is filled until a complete command
	// sequence is reached and then executed.
	cb []byte

	// os holds the state of each option.
	os optStates
}

// NewReadWriter allocates a new ReadWriter that intercepts and handles
// telnet negotiations and dispatches remaining data to rw.  Any options
// that are provided will be available for negotiation.
func NewReadWriter(rw io.ReadWriter, opts ...Option) *Ctx {
	t := &Ctx{rw: rw}

	t.os = make(optStates)
	for _, opt := range opts {
		t.os.store(optState{opt: opt})
	}

	return t
}

// AskHim asks him to enable or disable an option.
func (t Ctx) AskHim(opt Option, enable bool) error {
	if enable {
		return t.ask(do, opt)
	}

	return t.ask(dont, opt)
}

// AskUs asks if we can enable or disable an option.
func (t Ctx) AskUs(opt Option, enable bool) error {
	if enable {
		return t.ask(will, opt)
	}

	return t.ask(wont, opt)
}

// SendCmd sends a line mode command signal.
func (t Ctx) SendCmd(cmd Command) {
	t.rw.Write([]byte{byte(iac), byte(cmd)})
}

// SendParams sends option subnegotiation parameters.
func (t Ctx) SendParams(opt Option, params []byte) {
	t.rw.Write([]byte{byte(iac), byte(sb), opt.Byte()})
	t.rw.Write(params)
	t.rw.Write([]byte{byte(iac), byte(se)})
}
