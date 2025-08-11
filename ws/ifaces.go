package ws

import (
	"context"
	"net"
)

// Handler is a function used to handle WebSocket connections.
type Handler func(ctx context.Context, conn Conn)

// Conn represents a WebSocket connection interface.
// It provides methods to interact with the connection, such as reading and writing data,
// setting deadlines, and checking the connection status.
type Conn interface {
	// RemoteAddr returns the remote network address of the connection.
	RemoteAddr() net.Addr

	// Protocol returns the protocol used by the WebSocket connection.
	// Provided by the server during the handshake.
	Protocol() string

	// Alive checks if the WebSocket connection is still alive.
	Alive() bool

	// Read reads data from the WebSocket connection into the provided byte slice.
	Read([]byte) (int, error)

	// Write writes data to the WebSocket connection.
	Write([]byte) (int, error)

	// WriteClose writes a close frame to the WebSocket connection with the specified close code.
	WriteClose(WsCloseCode) error

	// Close closes the WebSocket connection.
	Close() error
}
