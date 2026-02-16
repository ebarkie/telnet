// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package telnet

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
)

// Errors.
var (
	ErrNegAskDenied = errors.New("ask violates let")
)

// negState is a RFC1143 option negotiation state.
type negState byte

const (
	nsNo         negState = iota // Disabled
	nsYes                        // Enabled
	nsWantNo                     // Negotiating for disable
	nsWantNoOpp                  // Want to enable but previous disable negotiation not complete
	nsWantYes                    // Negotiating for enable
	nsWantYesOpp                 // Want to disable but previous enable negotiation not complete
)

func (t *Ctx) indicate(cmd Command, code byte) {
	s := t.os.load(code)
	slog.Debug("indicating option", "cmd", cmd, "opt", s.opt)
	t.rw.Write([]byte{byte(iac), byte(cmd), code})
}

func (t *Ctx) ask(cmd Command, opt Option) (err error) {
	slog.Debug("asking option", "cmd", cmd, "opt", opt)
	s := t.os.load(opt.Byte())

	switch cmd {
	case will:
		// We are asking if we can enable an option.
		switch s.us {
		case nsNo:
			if s.opt.LetUs() {
				t.indicate(will, opt.Byte())
				s.us = nsWantYes
			} else {
				err = ErrNegAskDenied
			}
		case nsWantNo:
			s.us = nsWantNoOpp
		case nsWantYesOpp:
			s.us = nsWantYes
		}
	case wont:
		// We are indicating that we are disabling an option.
		switch s.us {
		case nsYes:
			t.indicate(wont, opt.Byte())
			s.us = nsWantNo
		case nsWantNoOpp:
			s.us = nsWantNo
		case nsWantYes:
			s.us = nsWantYesOpp
		}
	case do:
		// We are asking that he enable an option.
		switch s.him {
		case nsNo:
			if s.opt.LetHim() {
				t.indicate(do, opt.Byte())
				s.him = nsWantYes
			} else {
				err = ErrNegAskDenied
			}
		case nsWantNo:
			s.him = nsWantNoOpp
		case nsWantYesOpp:
			s.him = nsWantYes
		}
	case dont:
		// We are asking that he disable an option.
		switch s.him {
		case nsYes:
			t.indicate(dont, opt.Byte())
			s.him = nsWantNo
		case nsWantNoOpp:
			s.him = nsWantNo
		case nsWantYes:
			s.him = nsWantYesOpp
		}
	}

	t.os.store(s)

	return
}

func (t *Ctx) negotiate(cmd Command, code byte) (err error) {
	s := t.os.load(code)
	slog.Debug("received option", "cmd", cmd, "opt", s.opt)

	var callback func(*Ctx, bool)
	var enabled bool

	switch cmd {
	case will:
		// He is asking if he can enable an option or accepting our
		// request that he enable an option.
		switch s.him {
		case nsNo:
			if s.opt.LetHim() {
				t.indicate(do, code)
				s.him = nsYes

				slog.Debug("option enabled for him", "opt", s.opt)
				callback, enabled = s.opt.SetHim, true
			} else {
				t.indicate(dont, code)
			}
		case nsYes:
			// Ignore
		case nsWantNo:
			err = fmt.Errorf("%s option %s answered by %s", dont, s.opt, will)
			s.him = nsNo

			slog.Debug("option disabled for him", "opt", s.opt)
			callback, enabled = s.opt.SetHim, false
		case nsWantNoOpp:
			err = fmt.Errorf("%s option %s answered by %s", dont, s.opt, will)
			fallthrough
		case nsWantYes:
			s.him = nsYes

			slog.Debug("option enabled for him", "opt", s.opt)
			callback, enabled = s.opt.SetHim, true
		case nsWantYesOpp:
			t.indicate(dont, code)
			s.him = nsWantNo
		}
	case wont:
		// He is indicating that he is disabling an option, accepting our
		// request that he disable an option, or refusing our request for him
		// to enable an option.
		switch s.him {
		case nsNo:
			// Ignore
		case nsYes:
			t.indicate(dont, code)
			fallthrough
		case nsWantNo, nsWantYes, nsWantYesOpp:
			s.him = nsNo

			slog.Debug("option disabled for him", "opt", s.opt)
			callback, enabled = s.opt.SetHim, false
		case nsWantNoOpp:
			t.indicate(do, code)
			s.him = nsWantYes
		}
	case do:
		// He is accepting our request for us to enable an option or asking us
		// to enable an option.
		switch s.us {
		case nsNo:
			if s.opt.LetUs() {
				t.indicate(will, code)
				s.us = nsYes

				slog.Debug("option enabled for us", "opt", s.opt)
				callback, enabled = s.opt.SetUs, true
			} else {
				t.indicate(wont, code)
			}
		case nsYes:
			// Ignore
		case nsWantNo:
			err = fmt.Errorf("%s option %s answered by %s", wont, s.opt, do)
			s.us = nsNo

			slog.Debug("option disabled for us", "opt", s.opt)
			callback, enabled = s.opt.SetUs, false
		case nsWantNoOpp:
			err = fmt.Errorf("%s option %s answered by %s", wont, s.opt, do)
			fallthrough
		case nsWantYes:
			s.us = nsYes

			slog.Debug("option enabled for us", "opt", s.opt)
			callback, enabled = s.opt.SetUs, true
		case nsWantYesOpp:
			t.indicate(wont, code)
			s.us = nsWantNo

		}
	case dont:
		// He is refusing our request for us to enable an option or asking us
		// to disable an option.
		switch s.us {
		case nsNo:
			// Ignore
		case nsYes:
			t.indicate(wont, code)
			fallthrough
		case nsWantNo, nsWantYes, nsWantYesOpp:
			s.us = nsNo

			slog.Debug("option disabled for us", "opt", s.opt)
			callback, enabled = s.opt.SetUs, false
		case nsWantNoOpp:
			t.indicate(will, code)
			s.us = nsWantYes
		}
	}

	t.os.store(s)

	if callback != nil {
		t.mu.Unlock()
		callback(t, enabled)
		t.mu.Lock()
	}

	return
}

func (t *Ctx) subnegotiate(code byte, params []byte) {
	s := t.os.load(code)
	slog.Debug("subnegotiation", "opt", s.opt, "params", hex.Dump(params))

	t.mu.Unlock()
	s.opt.Params(t, params)
	t.mu.Lock()
}
