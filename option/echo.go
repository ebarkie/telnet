// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package option

import "gitlab.com/ebarkie/telnet"

// Echo is the RFC857 Telnet Echo Option.
type Echo struct {
	Us, Him bool
}

func (Echo) Byte() byte     { return 1 }
func (Echo) String() string { return "Echo" }

func (e Echo) LetHim() bool {
	// Us or him can be enabled but not both, otherwise
	// any character transmitted in either direction will be
	// echoed back and forth indefinitely.
	return !e.Us
}
func (e Echo) LetUs() bool { return !e.Him }

func (Echo) Params(tn *telnet.Ctx, params []byte) {}

func (e *Echo) SetHim(tn *telnet.Ctx, enabled bool) { e.Him = enabled }
func (e *Echo) SetUs(tn *telnet.Ctx, enabled bool)  { e.Us = enabled }
