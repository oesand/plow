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

	Alive() bool
	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error

	Read([]byte) (int, error)
	Write([]byte) error
	WriteClose(WsCloseCode) error
	Close() error
}
