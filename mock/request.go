package mock

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/internal/proxy"
	"github.com/oesand/plow/specs"
	"io"
	"net"
)

// DefaultRequest creates a new RequestBuilder with default values.
func DefaultRequest() *RequestBuilder {
	return &RequestBuilder{
		protoMajor: 1,
		protoMinor: 0,
		method:     specs.HttpMethodGet,
		remoteAddr: &net.TCPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 8080,
		},
		url: &specs.Url{
			Scheme: "http",
			Host:   "127.0.0.1",
			Port:   80,
			Path:   "/",
		},
		header: specs.NewHeader(),
	}
}

// RequestBuilder is used to build a plow.Request with customizable fields.
type RequestBuilder struct {
	protoMajor, protoMinor uint16

	hijacker   plow.HijackHandler
	remoteAddr net.Addr
	method     specs.HttpMethod
	url        *specs.Url
	header     *specs.Header
	body       io.Reader

	req *request
}

// Proto sets the protocol version for the request.
func (b *RequestBuilder) Proto(protoMajor, protoMinor uint16) *RequestBuilder {
	b.protoMajor = protoMajor
	b.protoMinor = protoMinor
	return b
}

// Addr sets the remote address for the request.
func (b *RequestBuilder) Addr(network, domain string, port int) *RequestBuilder {
	b.remoteAddr = &proxy.ResolvedAddr{
		Net:    network,
		Domain: domain,
		Port:   port,
	}
	return b
}

// Method sets the HTTP method for the request.
func (b *RequestBuilder) Method(method specs.HttpMethod) *RequestBuilder {
	b.method = method
	return b
}

// Url sets the URL for the request.
func (b *RequestBuilder) Url(url *specs.Url) *RequestBuilder {
	b.url = url
	return b
}

// Header returns the header for the request.
// If the header is nil, it initializes a new Header.
func (b *RequestBuilder) Header() *specs.Header {
	if b.header == nil {
		b.header = specs.NewHeader()
	}
	return b.header
}

// ConfHeader applies a configuration function to the request header.
// This allows for custom modifications to the header.
func (b *RequestBuilder) ConfHeader(conf func(*specs.Header)) *RequestBuilder {
	conf(b.Header())
	return b
}

// Hijacker returns the hijack handler for the request.
func (b *RequestBuilder) Hijacker() plow.HijackHandler {
	return b.hijacker
}

// Body sets the body for the request.
func (b *RequestBuilder) Body(body io.ReadCloser) *RequestBuilder {
	b.body = body
	return b
}

// Request returns a plow.Request based on the current state of the RequestBuilder.
func (b *RequestBuilder) Request() plow.Request {
	if b.header == nil {
		b.header = specs.NewHeader()
	}
	if b.req == nil {
		b.req = &request{b: b}
	}
	return b.req
}

type request struct {
	b *RequestBuilder
}

func (r request) ProtoVersion() (major, minor uint16) {
	return r.b.protoMajor, r.b.protoMinor
}

func (r request) RemoteAddr() net.Addr {
	return r.b.remoteAddr
}

func (r request) Hijack(handler plow.HijackHandler) {
	r.b.hijacker = handler
}

func (r request) Method() specs.HttpMethod {
	return r.b.method
}

func (r request) Url() *specs.Url {
	return r.b.url
}

func (r request) Header() *specs.Header {
	return r.b.header
}

func (r request) Body() io.Reader {
	return r.b.body
}
