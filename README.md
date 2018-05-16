# Telnet

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](http://choosealicense.com/licenses/mit/)
[![Build Status](https://travis-ci.org/ebarkie/telnet.svg?branch=master)](https://travis-ci.org/ebarkie/telnet)

Go package for creating a Telnet protocol ReadWriter, which would typically be
used to create a TCP telnet server.

Correctness is the primary focus and performance is secondary.

Options included:
* Echo us
* Suppress Go Ahead (SGA)
* Terminal-Type

Refer to the Option interface in [USAGE](USAGE.md) for information about
implementing other options.

Specifications:

| Document | Description                                            |
|----------|--------------------------------------------------------|
| RFC854   | Telnet Protocol Specification                          |
| RFC855   | Telnet Option Specifications                           |
| RFC857   | Telnet Echo Option                                     |
| RFC858   | Telnet Suppress Go Ahead Option                        |
| RFC1091  | Telnet Terminal-Type Option                            |
| RFC1143  | The Q Method of Implementing TELNET Option Negotiation |

## Installation

```
$ go get github.com/ebarkie/telnet{,/option}
```

## Usage

See [USAGE](USAGE.md).

## Example

```go
package main

import (
	"io"
	"log"
	"net"

	"github.com/ebarkie/telnet"
)

func serve(conn net.Conn) {
	defer func() {
		log.Printf("Connection from %s closed", conn.RemoteAddr())
		conn.Close()
	}()

	// Create telnet ReadWriter with no options.
	tn := telnet.NewReadWriter(conn)

	// Welcome banner.
	tn.Write([]byte("Welcome to a test telnet server!\r\n\r\n"))

	// Process input until connection is closed.
	buf := make([]byte, 1024)
	for {
		tn.Write([]byte("> "))
		n, err := tn.Read(buf)
		if err == io.EOF {
			return
		}
		log.Printf("Read '%s' {% [1]x} n=%d", buf[:n], n)
	}
}

func main() {
	// Create TCP listener.
	addr := net.JoinHostPort("127.0.0.1", "8023")
	log.Printf("Listening on %s", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, _ := l.Accept()
		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go serve(conn)
	}
}
```

## License

Copyright (c) 2018 Eric Barkie. All rights reserved.  
Use of this source code is governed by the MIT license
that can be found in the [LICENSE](LICENSE) file.
