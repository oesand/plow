package giglet

import (
	"context"
	"crypto/tls"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"net"
)

type HijackHandler = server.HijackHandler
type NextProtoHandler func(conn *tls.Conn)

// Handler serve http server requests and answer it
type Handler interface {
	Handle(ctx context.Context, request Request) Response
}

// RoundTripper send htto client requests and serve response
type RoundTripper interface {
	RoundTrip(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error)
}

// Dialer for start connection over network
type Dialer interface {
	Dial(ctx context.Context, network, address string) (net.Conn, error)
}

// TlsDialer dialer for tls connection and Handshake
type TlsDialer interface {
	Handshake(ctx context.Context, conn net.Conn, host string) (net.Conn, error)
}

// * Shorthand implementations *

// HandlerFunc shorthand implementation for Handler
type HandlerFunc func(ctx context.Context, request Request) Response

func (f HandlerFunc) Handle(ctx context.Context, request Request) Response {
	return f(ctx, request)
}

// RoundTripperFunc shorthand implementation for RoundTripper
type RoundTripperFunc func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error)

func (f RoundTripperFunc) RoundTrip(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
	return f(ctx, method, url, header, writer)
}

// DialerFunc shorthand implementation for Dialer
type DialerFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

// TlsDialerFunc shorthand implementation for TlsDialer
type TlsDialerFunc func(ctx context.Context, conn net.Conn, host string) (net.Conn, error)

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
