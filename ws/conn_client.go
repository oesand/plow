package ws

import (
	"bufio"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

func newClientConn(url specs.Url, conn net.Conn, rws *bufio.ReadWriter) *wsClientConn {
	return &wsClientConn{
		frameHandler: *newFrameHandler(rws, false),
		url:          &url,
		conn:         conn,
	}
}

type wsClientConn struct {
	frameHandler
	url    *specs.Url
	conn   net.Conn
	closed bool
}

func (conn *wsClientConn) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *wsClientConn) Url() *specs.Url {
	return conn.url
}

func (conn *wsClientConn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *wsClientConn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *wsClientConn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

func (conn *wsClientConn) Close() error {
	if conn.closed {
		return specs.ErrClosed
	}

	err := conn.WriteClose(CloseCodeNormal)
	err1 := conn.conn.Close()
	conn.closed = true

	stream.DefaultBufioReaderPool.Put(conn.rws.Reader)
	stream.DefaultBufioWriterPool.Put(conn.rws.Writer)

	if err != nil {
		return err
	}
	return err1
}
