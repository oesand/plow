package server

import (
	"context"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"time"
)

type HijackHandler func(conn net.Conn)

type HttpRequest struct {
	_ utils.NoCopy

	readTimeout time.Duration
	conn        net.Conn
	hijacker    HijackHandler
	context     context.Context

	protoMajor, protoMinor uint16
	method                 specs.HttpMethod
	url                    *specs.Url
	header                 *specs.Header

	BodyReader       io.Reader
	SelectedEncoding specs.ContentEncoding
}

func (req *HttpRequest) Context() context.Context {
	return req.context
}

func (req *HttpRequest) WithContext(context context.Context) {
	req.context = context
}

func (req *HttpRequest) ProtoVersion() (major, minor uint16) {
	return req.protoMajor, req.protoMinor
}

func (req *HttpRequest) RemoteAddr() net.Addr {
	return req.conn.RemoteAddr()
}

func (req *HttpRequest) Hijack(handler HijackHandler) {
	req.hijacker = handler
}

func (req *HttpRequest) Hijacker() HijackHandler {
	return req.hijacker
}

func (req *HttpRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *HttpRequest) Url() *specs.Url {
	return req.url
}

func (req *HttpRequest) Header() *specs.Header {
	return req.header
}

func (req *HttpRequest) Body() io.Reader {
	return req.BodyReader
}
