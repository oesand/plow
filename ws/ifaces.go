package ws

import (
	"context"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

// Handler is a function used to handle WebSocket connections.
type Handler func(ctx context.Context, conn Conn)

// Conn represents a WebSocket connection interface.
// It provides methods to interact with the connection, such as reading and writing data,
// setting deadlines, and checking the connection status.
type Conn interface {
	// RemoteAddr returns the remote network address of the connection.
	RemoteAddr() net.Addr

	// Url returns the URL associated with the WebSocket connection.
	Url() *specs.Url

	// Protocol returns the protocol used by the WebSocket connection.
	// Provided by the server during the handshake.
	Protocol() string

	// Alive checks if the WebSocket connection is still alive.
	Alive() bool

	// SetDeadline sets the deadline for the connection for all future read and write operations.
	SetDeadline(time.Time) error

	// SetReadDeadline sets the deadline for future read operations.
	SetReadDeadline(time.Time) error

	// SetWriteDeadline sets the deadline for future write operations.
	SetWriteDeadline(time.Time) error

	// Read reads data from the WebSocket connection into the provided byte slice.
	Read([]byte) (int, error)

	// Write writes data to the WebSocket connection.
	Write([]byte) (int, error)

	// WriteClose writes a close frame to the WebSocket connection with the specified close code.
	WriteClose(WsCloseCode) error

	// Close closes the WebSocket connection.
	Close() error
}
