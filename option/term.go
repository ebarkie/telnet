// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package option

import "github.com/ebarkie/telnet"

// Term is the RFC1091 Telnet Terminal-Type Option.
type Term struct {
	Him string
}

func (Term) Byte() byte     { return 24 }
func (Term) String() string { return "Terminal-Type" }

func (Term) LetHim() bool { return true }
func (Term) LetUs() bool  { return false }

func (t *Term) Params(tn *telnet.Ctx, params []byte) {
	const is byte = 0

	if len(params) > 1 && params[0] == is {
		t.Him = string(params[1:])
	}
}

func (t Term) SetHim(tn *telnet.Ctx, enabled bool) {
	const send byte = 1

	if enabled {
		tn.SendParams(&t, []byte{send})
	}
}

func (Term) SetUs(tn *telnet.Ctx, enabled bool) {}
