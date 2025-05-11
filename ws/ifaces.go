package ws

import (
	"context"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

type Conn interface {
	Context() context.Context
	WithContext(context context.Context)

	RemoteAddr() net.Addr
	Url() *specs.Url
	Header() *specs.Header

	Alive() bool
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error

	Read() (WebSocketFrame, []byte, error)
	Write(WebSocketFrame, []byte) error
	Close(WebSocketClose) error
}
