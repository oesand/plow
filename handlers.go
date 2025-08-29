package plow

import (
	"context"
	"crypto/tls"
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/specs"
	"net"
)

// HijackHandler function that allow an [Handler] to take over the connection.
type HijackHandler = server_ops.HijackHandler

// NextProtoHandler function to take over ownership
// of the provided TLS connection when an ALPN
// protocol upgrade has occurred.
type NextProtoHandler func(conn *tls.Conn)

// Handler serve HTTP [Server] requests and answer it
type Handler interface {
	Handle(ctx context.Context, request Request) Response
}

// ErrorHandler is an interface representing the ability to handle errors
// that occur during the processing of a request by a [Server].
type ErrorHandler interface {
	HandleError(ctx context.Context, conn net.Conn, err any)
}

// RoundTripper is an interface representing the ability to execute a
// single HTTP transaction, obtaining the [ClientResponse] for a
// given request parts [specs.HttpMethod], [specs.Url], [specs.Header] and
// optional [BodyWriter].
//
// A RoundTripper must be concurrent safe for use by multiple goroutines.
type RoundTripper interface {
	// RoundTrip executes a single HTTP transaction,
	// returning a [ClientResponse] for the provided request parts.
	RoundTrip(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error)
}

// Dialer is an interface representing the ability to connection over network to address.
type Dialer interface {
	// Dial connects to the address on the named network.
	Dial(ctx context.Context, network, address string) (net.Conn, error)
}

// TlsDialer is an interface representing the ability to dial tls connection and make Handshake.
type TlsDialer interface {
	// Handshake connects to the address on the named network.
	Handshake(ctx context.Context, conn net.Conn, host string) (net.Conn, error)
}

// * Shorthand implementations *

// HandlerFunc shorthand implementation for [Handler]
type HandlerFunc func(ctx context.Context, request Request) Response

// Handle triggers top level function [HandlerFunc]
func (f HandlerFunc) Handle(ctx context.Context, request Request) Response {
	return f(ctx, request)
}

// ErrorHandlerFunc shorthand implementation for [ErrorHandler]
type ErrorHandlerFunc func(ctx context.Context, conn net.Conn, err any)

// HandleError triggers top level function [ErrorHandlerFunc]
func (f ErrorHandlerFunc) HandleError(ctx context.Context, conn net.Conn, err any) {
	f(ctx, conn, err)
}

// RoundTripperFunc shorthand implementation for [RoundTripper]
type RoundTripperFunc func(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error)

// RoundTrip triggers top level function [RoundTripperFunc]
func (f RoundTripperFunc) RoundTrip(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
	return f(ctx, method, url, header, writer)
}

// DialerFunc shorthand implementation for [Dialer]
type DialerFunc func(ctx context.Context, network, address string) (net.Conn, error)

// Dial triggers top level function [DialerFunc]
func (f DialerFunc) Dial(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

// TlsDialerFunc shorthand implementation for [TlsDialer]
type TlsDialerFunc func(ctx context.Context, conn net.Conn, host string) (net.Conn, error)

// Handshake triggers top level function [TlsDialerFunc]
func (f TlsDialerFunc) Handshake(ctx context.Context, conn net.Conn, host string) (net.Conn, error) {
	return f(ctx, conn, host)
}

// FixedProxyUrl returns a proxy function (for use in a [Transport])
// that always returns the same URL.
func FixedProxyUrl(url *specs.Url) func(*specs.Url) (*specs.Url, error) {
	return func(*specs.Url) (*specs.Url, error) {
		return url, nil
	}
}
