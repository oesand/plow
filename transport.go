package plow

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/internal/catch"
	"github.com/oesand/plow/internal/client_ops"
	"github.com/oesand/plow/internal/encoding"
	"github.com/oesand/plow/internal/parsing"
	"github.com/oesand/plow/internal/proxy"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// DefaultTransport factory for creating [Transport]
// with optimal parameters for perfomance and safety
//
// Each call creates a new instance of [Transport]
func DefaultTransport() *Transport {
	return &Transport{
		HeadMaxLength:       2 << 20,  // 2 mb
		MaxBodySize:         10 << 20, // 10 mb
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		ProxyDialTimeout:    10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
}

// Transport is an implementation of [RoundTripper] that supports HTTP,
// HTTPS, and proxies like 'http', 'https', 'socks5' 'socks5h' (can be provided by Proxy).
//
// Transport is a low-level primitive for making HTTP and HTTPS requests.
// For high-level functionality, such as cookies and redirects, see [Client].
type Transport struct {
	_ internal.NoCopy

	// Dialer specifies the dialer for creating unencrypted TCP connections.
	// If Dialer is nil then the transport dials using package net
	Dialer Dialer

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	//
	// The proxy type is determined by the URL scheme. "http",
	// "https", "socks5", and "socks5h" are supported. If the scheme is empty,
	// "http" is assumed.
	// "socks5" is treated the same as "socks5h".
	//
	// If the proxy specs.Url contains a username & password subcomponents,
	// the proxy request will pass the username and password
	// in a `Proxy-Authorization` header.
	//
	// If Proxy is nil or returns a nil *specs.Url, no proxy is used.
	Proxy func(url *specs.Url) (*specs.Url, error)

	// ProxyDialTimeout specifies the maximum amount of time to
	// wait for a Proxy establish connection.
	//
	// If zero there is no timeout.
	ProxyDialTimeout time.Duration

	// TLSDialer specifies an optional dial function for creating
	// TLS connections and gone handshake for non-proxied HTTPS requests.
	//
	// If TLSDialer is nil will be used [tls.Client].
	//
	// If TLSDialer is set, it is assumed
	// that the returned net.Conn has already gone through the TLS handshake.
	TLSDialer TlsDialer

	// TLSClientConfig specifies the TLS configuration
	// to use with tls.Client.
	//
	// If nil, the default configuration is used.
	// If non-nil, HTTP/2 support may not be enabled by default.
	TLSConfig *tls.Config

	// TLSHandshakeTimeout specifies the maximum amount of time to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	// ReadTimeout is the maximum duration for server the entire
	// response, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the request. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

	// ReadLineMaxLength maximum size in bytes
	// to read lines in the response
	// such as headers and headlines
	//
	// The client returns specs.ErrTooLarge if this limit is greater than 0
	// and response lines is greater than the limit.
	//
	// If zero there is no limit
	ReadLineMaxLength int64

	// HeadMaxLength maximum size in bytes
	// to read headline and headers together
	//
	// The client returns specs.ErrTooLarge if this limit is greater than 0
	// and response header size is greater than the limit.
	//
	// If zero there is no limit
	HeadMaxLength int64

	// MaxBodySize maximum size in bytes
	// to read response body size.
	//
	// The client returns specs.ErrTooLarge if this limit is greater than 0
	// and response body is greater than the limit.
	//
	// By default, response body size is unlimited.
	MaxBodySize int64
}

// RoundTrip implements the [RoundTripper] interface.
func (transport *Transport) RoundTrip(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
	if ctx == nil {
		panic("plow: nil context pointer")
	}
	if header == nil {
		panic("plow: nil header pointer")
	}
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid request method '%s'", method)
	}
	if !(url.Scheme == "http" || url.Scheme == "https") {
		return nil, fmt.Errorf("invalid request url '%s' scheme", url.Scheme)
	}
	if url.Host == "" {
		return nil, fmt.Errorf("invalid request url '%s' host", url.Host)
	}

	if url.Port == 0 {
		switch url.Scheme {
		case "http":
			url.Port = 80
		case "https":
			url.Port = 443
		}
	}

	var err error
	var proxyUrl *specs.Url
	if transport.Proxy != nil {
		proxyUrl, err = transport.Proxy(url)
		if err != nil {
			return nil, err
		}
	}

	host := client_ops.IdnaHost(url.Host)
	if proxyUrl != nil {
		if proxyUrl.Host == "" {
			return nil, fmt.Errorf("invalid proxy url '%s' host", url.Host)
		}

		if proxyUrl.Scheme == "http" && url.Scheme == "https" {
			proxyUrl.Scheme = url.Scheme
		}

		switch proxyUrl.Scheme {
		case "http":
			header.Set("Host", client_ops.HostHeader(host, url.Port, true))
			if proxyUrl.Username != "" {
				proxy.WithAuthHeader(header, proxyUrl.Username, proxyUrl.Password)
			}
		case "https", "socks5", "socks5h":
			header.Set("Host", client_ops.HostHeader(host, url.Port, false))
		default:
			return nil, fmt.Errorf("unsupported proxy '%s' scheme", url.Scheme)
		}

		proxyUrl.Host = client_ops.IdnaHost(proxyUrl.Host)
		if proxyUrl.Port == 0 {
			proxyUrl.Port = proxy.SchemeDefaultPortMap[proxyUrl.Scheme]
		}
	} else {
		header.Set("Host", client_ops.HostHeader(host, url.Port, false))
	}

	if !header.Has("Accept") {
		header.Set("Accept", "*/*")
	}

	if !header.Has("Accept-Encoding") &&
		!header.Has("Range") &&
		method != specs.HttpMethodHead {
		header.Set("Accept-Encoding", encoding.DefaultAcceptEncoding)
	}

	if url.Username != "" && !header.Has("Authorization") {
		header.Set("Authorization", specs.BasicAuthHeader(url.Username, url.Password))
	}

	if !header.Has("Connection") {
		header.Set("Connection", "close")
	}

	var isChunked bool
	if te, has := header.TryGet("Transfer-Encoding"); has {
		switch te {
		case "chunked":
			isChunked = true
		default:
			return nil, specs.ErrUnknownTransferEncoding
		}
	}

	mustWriteBody := method.IsPostable() && writer != nil

	if !isChunked && mustWriteBody {
		contentLength := writer.ContentLength()
		if contentLength > 0 {
			header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
		}
	}

	var conn net.Conn
	if proxyUrl != nil {
		conn, err = transport.dial(ctx, proxyUrl.Host, proxyUrl.Port)
		if err != nil {
			return nil, catch.TryWrapOpErr("dial", err)
		}

		var proxyCreds *proxy.Creds
		if proxyUrl.Username != "" {
			proxyCreds = &proxy.Creds{Username: proxyUrl.Username, Password: proxyUrl.Password}
		}

		err = transport.dialProxy(ctx, conn, proxyUrl.Scheme, host, url.Port, proxyCreds)
		if err != nil {
			conn.Close()
			return nil, catch.TryWrapOpErr("proxy", err)
		}
	} else {
		conn, err = transport.dial(ctx, host, url.Port)
		if err != nil {
			return nil, catch.TryWrapOpErr("dial", err)
		}
	}

	closeConn, includeClose, cancelCloseConn := internal.CancellableDefer(func() {
		conn.Close()
	})
	defer closeConn()

	if url.Scheme == "https" {
		var tlsConn net.Conn
		tlsConn, err = transport.dialTls(ctx, conn, host)
		if err != nil {
			return nil, catch.TryWrapOpErr("tls", err)
		}
		conn = tlsConn
	}

	if transport.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(transport.WriteTimeout))
		defer conn.SetWriteDeadline(time.Time{})
	}

	_, err = client_ops.WriteRequestHead(conn, method, url.Path, url.Query, header)

	if err == nil {
		err = ctx.Err()
	}
	if err = catch.CatchCommonErr(err); err != nil {
		return nil, catch.TryWrapOpErr("write", err)
	}

	// Expect 100 Continue support
	expectContinue := mustWriteBody && strings.EqualFold(header.Get("Expect"), "100-continue")
	if expectContinue {
		goto reading
	}

writeBody:
	if mustWriteBody {
		if isChunked {
			chunkedWriter := encoding.NewChunkedWriter(conn)
			err = writer.WriteBody(chunkedWriter)
			chunkedWriter.Close()
		} else {
			err = writer.WriteBody(conn)
		}

		if err == nil {
			err = ctx.Err()
		}
		if err != nil {
			return nil, catch.CatchCommonErr(err)
		}
	}

reading:
	if transport.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(transport.ReadTimeout))
	}

	bufioReader := stream.DefaultBufioReaderPool.Get(conn)
	includeClose(func() {
		stream.DefaultBufioReaderPool.Put(bufioReader)
	})

	resp, err := client_ops.ReadResponse(ctx, bufioReader, transport.ReadLineMaxLength, transport.HeadMaxLength)

	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return nil, catch.CatchCommonErr(err)
	}

	if expectContinue && resp.StatusCode() == specs.StatusCodeContinue {
		expectContinue = false
		goto writeBody
	}

	hijacker, hasHijacker := ctx.Value(transportHijackerKey).(*TransportHijacker)

	if !method.IsReplyable() || !resp.StatusCode().IsReplyable() {
		if hasHijacker && !strings.EqualFold(header.Get("Connection"), "close") {
			hijacker.Conn = conn
			cancelCloseConn()
		}
	} else {
		contentEncoding := resp.Header().Get("Content-Encoding")
		if contentEncoding != "" && !encoding.IsKnownEncoding(contentEncoding) {
			return nil, specs.ErrUnknownContentEncoding
		}

		var contentLength int64
		isChunked, contentLength, err = parsing.ParseContentLength(resp.Header())
		if err != nil {
			if errors.Is(err, parsing.ErrParsing) {
				// Fail to parse Content-Length
				stream.DefaultBufioReaderPool.Put(bufioReader)
				if hasHijacker {
					return nil, errors.New("cannot parse Content-Length value")
				}
			} else {
				return nil, err
			}
		} else if isChunked || contentLength > 0 {
			if maxSize := transport.MaxBodySize; maxSize > 0 {
				if isChunked {
					contentLength = maxSize
				} else if contentLength > maxSize {
					return nil, specs.ErrTooLarge
				}
			}

			encodingReader, err := encoding.NewReader(isChunked, contentEncoding, bufioReader)
			if err != nil {
				return nil, err
			}

			var bodyReader io.Reader = encodingReader

			if contentLength > 0 {
				bodyReader = io.LimitReader(bodyReader, contentLength)
			}

			cancelCloseConn()
			resp.Reader = internal.ReadCloser(bodyReader, internal.CloserFunc(func() error {
				stream.DefaultBufioReaderPool.Put(bufioReader)

				err := encodingReader.Close()
				err1 := conn.Close()
				if err != nil {
					return err
				}
				return err1
			}))
		}

		if hasHijacker {
			hijacker.Conn = conn
			cancelCloseConn()
		}
	}

	return resp, nil
}

func (transport *Transport) dial(ctx context.Context, host string, port uint16) (net.Conn, error) {
	address := client_ops.HostPort(host, port)

	var conn net.Conn
	var err error
	if transport.Dialer != nil {
		conn, err = transport.Dialer.Dial(ctx, "tcp", address)
	} else {
		conn, err = defaultDialer.DialContext(ctx, "tcp", address)
	}

	return conn, catch.CatchCommonErr(err)
}

func (transport *Transport) dialProxy(ctx context.Context, conn net.Conn, scheme, host string, port uint16, creds *proxy.Creds) error {
	if scheme == "http" {
		return nil
	}
	if transport.ProxyDialTimeout > 0 {
		conn.SetDeadline(time.Now().Add(transport.ProxyDialTimeout))
		defer conn.SetDeadline(time.Time{})
	}
	return catch.CallWithTimeoutContextErr(ctx, transport.ProxyDialTimeout, func(ctx context.Context) error {
		var err error
		switch scheme {
		case "https":
			err = proxy.DialHttps(conn, host, port, creds)
		case "socks5", "socks5h":
			_, err = proxy.DialSocks5(conn, host, port, creds)
		default:
			panic(fmt.Sprintf("plow: not implemented proxy '%s' scheme dialer", scheme))
		}
		return err
	})
}

func (transport *Transport) dialTls(ctx context.Context, conn net.Conn, host string) (net.Conn, error) {
	return catch.CallWithTimeoutContext(ctx, transport.TLSHandshakeTimeout, func(ctx context.Context) (net.Conn, error) {
		if transport.TLSDialer != nil {
			return transport.TLSDialer.Handshake(ctx, conn, host)
		} else {
			var tlsCfg *tls.Config

			if transport.TLSConfig == nil {
				tlsCfg = &tls.Config{}
			} else {
				tlsCfg = transport.TLSConfig.Clone()
			}

			if tlsCfg.ServerName == "" {
				tlsCfg.ServerName = host
			}

			tlsConn := tls.Client(conn, tlsCfg)
			err := tlsConn.HandshakeContext(ctx)
			if err != nil {
				return nil, err
			}
			return tlsConn, nil
		}
	})
}

var transportHijackerKey = internal.FlagKey{Key: "transport.hijacker.key"}

// WithTransportHijacker returns a copy of [context.Context] in which
// the stored TransportHijacker.
//
// If hijacker stored Transport will not do anything else
// with the connection when HTTP transaction ends.
// If ClientResponse has Body and you close it,
// it will also close net.Conn, keep this in mind.
//
// Used only with Transport or Client for intercept connection.
func WithTransportHijacker(ctx context.Context) (*TransportHijacker, context.Context) {
	if hijacker, has := ctx.Value(transportHijackerKey).(*TransportHijacker); has {
		return hijacker, nil
	}
	hijacker := new(TransportHijacker)
	return hijacker, context.WithValue(ctx, transportHijackerKey, hijacker)
}

// TransportHijacker container for store intercepted Transport connection.
//
// Creates only by WithTransportHijacker
type TransportHijacker struct {
	// Conn intercepted connection from Transport
	//
	// Can be nil if connection not stored
	Conn net.Conn
}
