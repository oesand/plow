package plow

import (
	"io"
	"net"

	"github.com/oesand/plow/specs"
)

// Request is an interface for an HTTP request received by the [Server].
type Request interface {
	// ProtoVersion specifies protocol version of incoming server request.
	ProtoVersion() (major, minor uint16)

	// RemoteAddr specifies client [net.Addr] of incoming server request.
	RemoteAddr() net.Addr

	// Hijack lets the handler take over the connection.
	// After a call to Hijack the HTTP server library
	// will not do anything else with the connection when HTTP transaction ends.
	Hijack(handler HijackHandler)

	// Method specifies the HTTP method (GET, POST, PUT, etc.)
	// of incoming server request.
	Method() specs.HttpMethod

	// Url specifies the URL of incoming server request.
	Url() *specs.Url

	// Header contains the header fields and cookies
	// of incoming server request.
	Header() *specs.Header

	// Body specifies [io.Reader] of incoming server request.
	//
	// if request body not provided return nil.
	Body() io.Reader
}

// Response is an interface for the HTTP response sent by the [Server].
type Response interface {
	// StatusCode specifies [specs.StatusCode] to be sent by the server in HTTP request.
	StatusCode() specs.StatusCode

	// Header contains the response header fields and cookies
	// to be sent by the server in HTTP response.
	Header() *specs.Header
}

// ClientRequest is an interface for the HTTP response sent
// by the [Client] and [RoundTripper] (such as [Transport]).
type ClientRequest interface {
	// Method specifies the HTTP method (GET, POST, PUT, etc.)
	// to be sent by the client in HTTP request.
	Method() specs.HttpMethod

	// Url specifies the URL for client requests.
	Url() *specs.Url

	// Header contains the request header fields and cookies
	// to be sent by the client in HTTP request.
	Header() *specs.Header
}

// ClientResponse is an interface for an HTTP request received
// by the [Client] and [RoundTripper] (such as [Transport]).
type ClientResponse interface {
	// StatusCode specifies [specs.StatusCode] which
	// is received by the client.
	StatusCode() specs.StatusCode

	// Header contains the request header fields and cookies
	// which is received by the client.
	Header() *specs.Header

	// Body specifies [io.ReadCloser] response body
	// which is received by the client.
	//
	// if response body not provided return nil.
	Body() io.ReadCloser
}

// BodyWriter is an interface representing the ability
// to send bytes of the HTTP server response body or client request body.
type BodyWriter interface {
	// WriteBody function allows you to send a content body to a specific node.
	WriteBody(io.Writer) error

	// ContentLength specifies size of the body
	// which can be passed in the 'Content-Length' header.
	ContentLength() int64
}

// MarshallResponse is an interface that combines [Response] and [BodyWriter] capabilities
// with the ability to provide an instance of the underlying response type.
type MarshallResponse interface {
	Response
	BodyWriter
	Instance() any
}
