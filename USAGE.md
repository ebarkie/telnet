# telnet

```go
import "gitlab.com/ebarkie/telnet"
```

Package telnet implements the RFC854 Telnet Protocol Specification, as well as:

    RFC855  Telnet Option Specifications
    RFC1143 The Q Method of Implementing TELNET Option Negotiation

## Usage

```go
var (
	Trace = log.New(ioutil.Discard, "[TRCE]", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	Debug = log.New(ioutil.Discard, "[DBUG]", log.LstdFlags|log.Lshortfile)
	Error = log.New(ioutil.Discard, "[ERRO]", log.LstdFlags|log.Lshortfile)
)
```
Loggers.

```go
var (
	ErrNegAskDenied = errors.New("Ask violates let")
)
```
Errors.

#### type Command

```go
type Command byte
```

Command is a telnet command.

```go
const (
	EOF Command = 236 + iota // End of File character
	SP                       // Suspend Process
	AP                       // Abort Process
	EOR                      // End Of Record

	NOP // No Operation
	DM  // Data Mark
	BRK // Break
	IP  // Interrupt Process
	AO  // Abort Output
	AYT // Are You There
	EC  // Erase Character
	EL  // Erase Line
	GA  // Go Ahead

)
```
RFC854 telnet commands.

#### func (Command) String

```go
func (i Command) String() string
```

#### type Ctx

```go
type Ctx struct {
}
```

Ctx is a telnet context.

#### func  NewReadWriter

```go
func NewReadWriter(rw io.ReadWriter, opts ...Option) *Ctx
```
NewReadWriter allocates a new ReadWriter that intercepts and handles telnet
negotiations and dispatches remaining data to rw. Any options that are provided
will be available for negotiation.

#### func (Ctx) AskHim

```go
func (t Ctx) AskHim(opt Option, enable bool) error
```
AskHim asks him to enable or disable an option.

#### func (Ctx) AskUs

```go
func (t Ctx) AskUs(opt Option, enable bool) error
```
AskUs asks if we can enable or disable an option.

#### func (*Ctx) Read

```go
func (t *Ctx) Read(b []byte) (n int, err error)
```
Read is a telnet Reader.

The Reader provided by the client is read and command sequences are parsed and
executed. If a non-empty buffer is passed then data will be passed as-is.

#### func (Ctx) SendCmd

```go
func (t Ctx) SendCmd(cmd Command)
```
SendCmd sends a line mode command signal.

#### func (Ctx) SendParams

```go
func (t Ctx) SendParams(opt Option, params []byte)
```
SendParams sends option subnegotiation parameters.

#### func (Ctx) Write

```go
func (t Ctx) Write(b []byte) (int, error)
```
Write is a telnet Writer.

Any Interpret as Command bytes are escaped and the result is written using the
Writer provided by the client.

#### type Option

```go
type Option interface {
	// Byte returns the byte code of the option.
	Byte() byte
	// String is the text name of the option (for debugging).
	String() string

	// LetHim indicates if he is allowed to enable the option.
	LetHim() bool
	// LetUs indicates if we are willing to enable the option.
	LetUs() bool

	// Params is called when subnegotiation parameters are received.
	Params(tn *Ctx, params []byte)

	// SetHim sets the option state for him.
	SetHim(tn *Ctx, enabled bool)
	// SetUs sets the option state for us.
	SetUs(tn *Ctx, enabled bool)
}
```

Option is an interface for implementing telnet options.
