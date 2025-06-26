package ws

import (
	"bufio"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

func newServerConn(req giglet.Request, conn net.Conn, rws *bufio.ReadWriter) *wsServerConn {
	return &wsServerConn{
		frameHandler: *newFrameHandler(rws, true),
		request:      req,
		conn:         conn,
	}
}

type wsServerConn struct {
	frameHandler
	request giglet.Request
	conn    net.Conn
}

func (conn *wsServerConn) RemoteAddr() net.Addr {
	return conn.request.RemoteAddr()
}

func (conn *wsServerConn) Url() *specs.Url {
	return conn.request.Url()
}

func (conn *wsServerConn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *wsServerConn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *wsServerConn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

func (conn *wsServerConn) Close() error {
	err := conn.WriteClose(CloseCodeNormal)
	err1 := conn.conn.Close()
	if err != nil {
		return err
	}
	return err1
}
