package server_ops

import (
	"context"
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/specs"
	"io"
	"net"
)

type HijackHandler func(ctx context.Context, conn net.Conn)

type HttpRequest struct {
	_ internal.NoCopy

	hijacker HijackHandler

	protoMajor, protoMinor uint16
	remoteAddr             net.Addr
	method                 specs.HttpMethod
	url                    *specs.Url
	header                 *specs.Header

	BodyReader       io.Reader
	Chunked          bool
	SelectedEncoding string
}

func (req *HttpRequest) ProtoVersion() (major, minor uint16) {
	return req.protoMajor, req.protoMinor
}

func (req *HttpRequest) RemoteAddr() net.Addr {
	return req.remoteAddr
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
