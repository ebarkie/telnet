// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package option

import "github.com/ebarkie/telnet"

// SGA is the RFC858 Telnet Suppress Go Ahead Option.
type SGA struct{}

func (SGA) Byte() byte     { return 3 }
func (SGA) String() string { return "Suppress Go Ahead" }

func (SGA) LetHim() bool { return true }
func (SGA) LetUs() bool  { return true }

func (SGA) Params(tn *telnet.Ctx, params []byte) {}

func (SGA) SetHim(tn *telnet.Ctx, enabled bool) {}
func (SGA) SetUs(tn *telnet.Ctx, enabled bool)  {}
