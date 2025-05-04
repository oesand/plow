package giglet

import (
	"context"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
)

type Request interface {
	Context() context.Context
	WithContext(context context.Context)

	ProtoVersion() (major, minor uint16)
	RemoteAddr() net.Addr
	Hijack(handler HijackHandler)

	Method() specs.HttpMethod
	Url() *specs.Url
	Header() *specs.Header

	Body() io.Reader
}

type Response interface {
	StatusCode() specs.StatusCode
	SetStatusCode(specs.StatusCode)
	Header() *specs.Header
}

type ClientRequest interface {
	Method() specs.HttpMethod
	Url() specs.Url
	Header() *specs.Header
}

type ClientResponse interface {
	StatusCode() specs.StatusCode
	Header() *specs.Header
	Body() io.ReadCloser
}

type BodyWriter interface {
	WriteBody(io.Writer) error
}
