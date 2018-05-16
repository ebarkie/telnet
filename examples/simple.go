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
