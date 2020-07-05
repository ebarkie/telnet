package main

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/ebarkie/telnet"
	"github.com/ebarkie/telnet/option"
)

// negotiate creates a telnet ReadWriter, asks for character mode options, and blocks until
// the options are confirmed or the wait duration passes.
func negotiate(conn net.Conn, wait time.Duration, opts ...telnet.Option) (*telnet.Ctx, bool) {
	// Create telnet ReadWriter with options.
	echo := &option.Echo{}
	sga := &option.SGA{}
	tn := telnet.NewReadWriter(conn, append(opts, echo, sga)...)

	// Send option negotiations.
	tn.AskUs(sga, true)
	tn.AskUs(echo, true)

	// Wait for echo confirmation.
	conn.SetReadDeadline(time.Now().Add(wait))
	defer conn.SetReadDeadline(time.Time{})

	var err error
	for err == nil {
		_, err = tn.Read([]byte{})
		if echo.Us {
			break
		}
	}

	return tn, echo.Us
}

func serve(conn net.Conn) {
	defer conn.Close()
	defer log.Printf("Connection from %s closed", conn.RemoteAddr())

	// Create telnet ReadWriter with character mode enabled.
	term := &option.Term{}
	tn, ok := negotiate(conn, 2*time.Second, term)
	if !ok {
		tn.Write([]byte("Protocol negotiation failed.\r\n"))
		return
	}
	tn.AskHim(term, true)

	// Welcome banner and prompt.
	tn.Write([]byte("Welcome to a test telnet server!\r\n"))
	tn.Write([]byte("\r\n> "))

	// Process input until connection is closed.
	buf := make([]byte, 1024)
	var nt int // Number typed
	for {
		n, err := tn.Read(buf)
		if err == io.EOF {
			return
		}
		log.Printf("Read '%s' {% [1]x} n=%d", buf[:n], n)

		for i := 0; i < n; i++ {
			switch buf[i] {
			case 0x00, 0x0a:
				// A null or newline indicates end of line
				tn.Write([]byte("\r\n> "))
				nt = 0
			case 0x0d:
				// Ignore carriage returns
			case 0x08, 0x7f:
				// Backspace (^H or ^?)
				if nt > 0 {
					tn.Write([]byte{0x08, 0x20, 0x08})
					nt--
				}
			default:
				tn.Write([]byte{buf[i]})
				nt++
			}
		}
	}
}

func main() {
	// Attach loggers to standard out for debugging.
	for _, l := range []*log.Logger{
		telnet.Trace,
		telnet.Debug,
		telnet.Error,
	} {
		l.SetOutput(os.Stdout)
	}

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
