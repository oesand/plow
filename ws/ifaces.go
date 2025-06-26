package ws

import (
	"context"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

type Handler func(ctx context.Context, conn Conn)

type Conn interface {
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
