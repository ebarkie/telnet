// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package telnet

import "fmt"

// noOpt is used when unknown options are encountered during negotiation.
type noOpt struct {
	Code byte
}

func (n noOpt) Byte() byte     { return n.Code }
func (n noOpt) String() string { return fmt.Sprintf("Unknown-%d", n.Code) }

func (noOpt) LetHim() bool { return false }
func (noOpt) LetUs() bool  { return false }

func (noOpt) Params(tn *Ctx, params []byte) {}

func (noOpt) SetHim(tn *Ctx, enabled bool) {}
func (noOpt) SetUs(tn *Ctx, enabled bool)  {}
