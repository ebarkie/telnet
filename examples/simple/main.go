package main

import (
	"io"
	"log/slog"
	"net"

	"github.com/ebarkie/telnet"
)

func serve(conn net.Conn) {
	defer conn.Close()
	defer slog.Info("connection closed", "addr", conn.RemoteAddr())

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
		slog.Info("read", "data", buf[:n], "n", n)
	}
}

func main() {
	// Create TCP listener.
	addr := net.JoinHostPort("127.0.0.1", "8023")
	slog.Info("listening", "addr", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}
		slog.Info("accepted connection", "addr", conn.RemoteAddr())
		go serve(conn)
	}
}
