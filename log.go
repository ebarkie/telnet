// Copyright (c) 2018 Eric Barkie. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package telnet

import (
	"io/ioutil"
	"log"
)

// Loggers.
var (
	Trace = log.New(ioutil.Discard, "[TRCE]", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	Debug = log.New(ioutil.Discard, "[DBUG]", log.LstdFlags|log.Lshortfile)
	Error = log.New(ioutil.Discard, "[ERRO]", log.LstdFlags|log.Lshortfile)
)
