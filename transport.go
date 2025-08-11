package giglet

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/encoding"
	"github.com/oesand/giglet/internal/proxy"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"time"
)

// DefaultTransport factory for creating [Transport]
// with optimal parameters for perfomance and safety
//
// Each call creates a new instance of [Transport]
func DefaultTransport() *Transport {
	return &Transport{
		ReadLineMaxLength:   1024,
		HeadMaxLength:       8 * 1024,
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

	// ReadTimeout is the maximum duration for server the entire
	// response, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the request. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration
}

// RoundTrip implements the [RoundTripper] interface.
func (transport *Transport) RoundTrip(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
	if ctx == nil {
		panic("nil context pointer")
	}
	if header == nil {
		panic("nil header pointer")
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
		proxyUrl, err = transport.Proxy(&url)
		if err != nil {
			return nil, err
		}
	}

	host := client.IdnaHost(url.Host)
	if proxyUrl != nil {
		if proxyUrl.Host == "" {
			return nil, fmt.Errorf("invalid proxy url '%s' host", url.Host)
		}

		if proxyUrl.Scheme == "http" && url.Scheme == "https" {
			proxyUrl.Scheme = url.Scheme
		}

		switch proxyUrl.Scheme {
		case "http":
			header.Set("Host", client.HostHeader(host, url.Port, true))
			if proxyUrl.Username != "" {
				proxy.WithAuthHeader(header, proxyUrl.Username, proxyUrl.Password)
			}
		case "https", "socks5", "socks5h":
			header.Set("Host", client.HostHeader(host, url.Port, false))
		default:
			return nil, fmt.Errorf("unsupported proxy '%s' scheme", url.Scheme)
		}

		proxyUrl.Host = client.IdnaHost(proxyUrl.Host)
		if proxyUrl.Port == 0 {
			proxyUrl.Port = proxy.SchemeDefaultPortMap[proxyUrl.Scheme]
		}
	} else {
		header.Set("Host", client.HostHeader(host, url.Port, false))
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

	isChunked, err := encoding.IsChunkedTransfer(header)
	if err != nil {
		return nil, err
	}

	if !isChunked && writer != nil {
		contentLength := writer.ContentLength()
		if contentLength > 0 {
			header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
		}
	}

	var conn net.Conn
	if proxyUrl != nil {
		conn, err = transport.dial(ctx, proxyUrl.Host, proxyUrl.Port)
		if err = catch.CatchCommonErr(err); err != nil {
			return nil, &specs.GigletError{
				Op:  "dial",
				Err: err,
			}
		}

		var proxyCreds *proxy.Creds
		if proxyUrl.Username != "" {
			proxyCreds = &proxy.Creds{Username: proxyUrl.Username, Password: proxyUrl.Password}
		}

		err = transport.dialProxy(ctx, conn, proxyUrl.Scheme, host, url.Port, proxyCreds)
		if err = catch.CatchCommonErr(err); err != nil {
			conn.Close()
			return nil, &specs.GigletError{
				Op:  "proxy",
				Err: err,
			}
		}
	} else {
		conn, err = transport.dial(ctx, host, url.Port)
		if err = catch.CatchCommonErr(err); err != nil {
			return nil, &specs.GigletError{
				Op:  "dial",
				Err: err,
			}
		}
	}

	if url.Scheme == "https" {
		conn, err = transport.dialTls(ctx, conn, host)
		if err = catch.CatchCommonErr(err); err != nil {
			conn.Close()
			return nil, &specs.GigletError{
				Op:  "tls",
				Err: err,
			}
		}
	}

	if transport.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(transport.WriteTimeout))
	}

	_, err = client.WriteRequestHead(conn, method, url.Path, url.Query, header)

	if err = catch.CatchCommonErr(err); err != nil {
		conn.Close()
		return nil, &specs.GigletError{
			Op:  "write",
			Err: err,
		}
	} else if err = catch.CatchContextCancel(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	if method.IsPostable() && writer != nil {
		if isChunked {
			chunkedWriter := encoding.NewChunkedWriter(conn)
			err = writer.WriteBody(chunkedWriter)
			chunkedWriter.Close()
		} else {
			err = writer.WriteBody(conn)
		}

		if err != nil {
			conn.Close()
			return nil, catch.CatchCommonErr(err)
		}

		if err = catch.CatchContextCancel(ctx); err != nil {
			conn.Close()
			return nil, err
		}
	}

	if transport.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Time{})
	}

	if transport.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(transport.ReadTimeout))
	}

	headerReader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(headerReader)

	resp, err := client.ReadResponse(ctx, headerReader, transport.ReadLineMaxLength, transport.HeadMaxLength)

	err = catch.CatchCommonErr(err)
	if err == nil {
		err = catch.CatchContextCancel(ctx)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}

	hijacker, hasHijacker := ctx.Value(transportHijackerKey).(*TransportHijacker)

	if !method.IsReplyable() || !resp.StatusCode().IsReplyable() {
		if hasHijacker {
			if header.Get("Connection") == "close" {
				conn.Close()
			} else {
				hijacker.Conn = conn
			}
		} else {
			conn.Close()
		}
	} else {
		contentEncoding := resp.Header().Get("Content-Encoding")
		if contentEncoding != "" && !encoding.IsKnownEncoding(contentEncoding) {
			return nil, specs.ErrUnknownContentEncoding
		}

		extraBuffered, _ := headerReader.Peek(headerReader.Buffered())
		reader := io.MultiReader(bytes.NewReader(extraBuffered), conn)

		isChunked, err = encoding.IsChunkedTransfer(resp.Header())
		if err != nil {
			conn.Close()
			return nil, err
		}

		if isChunked {
			if transport.MaxBodySize > 0 {
				reader = io.LimitReader(reader, transport.MaxBodySize)
			}
		} else {
			var contentLength int64
			if contentLengthString := resp.Header().Get("Content-Length"); contentLengthString != "" {
				contentLength, err = strconv.ParseInt(contentLengthString, 10, 64)
				if err == nil {
					if contentLength <= 0 {
						reader = nil
						conn.Close()
					} else if transport.MaxBodySize > 0 && contentLength >= transport.MaxBodySize {
						return nil, specs.ErrTooLarge
					}
				}
			}

			if reader != nil {
				if contentLength <= 0 && transport.MaxBodySize > 0 {
					contentLength = transport.MaxBodySize
				}

				if contentLength > 0 {
					reader = io.LimitReader(reader, contentLength)
				}
			}
		}

		if reader != nil {
			var readerClosers internal.SeqCloser

			readerClosers.Add(conn)

			if isChunked {
				reader = httputil.NewChunkedReader(reader)
			}

			if contentEncoding != "" {
				var encodingReader io.ReadCloser
				encodingReader, err = encoding.NewReader(contentEncoding, reader)
				if err != nil {
					conn.Close()
					return nil, err
				}
				readerClosers.Add(encodingReader)
				reader = encodingReader
			}

			resp.Reader = internal.ReadCloser(reader, &readerClosers)

			if err = catch.CatchContextCancel(ctx); err != nil {
				conn.Close()
				return nil, err
			}
		}

		if hasHijacker {
			hijacker.Conn = conn
		}
	}

	return resp, nil
}

func (transport *Transport) dial(ctx context.Context, host string, port uint16) (net.Conn, error) {
	address := client.HostPort(host, port)

	var conn net.Conn
	var err error
	if transport.Dialer != nil {
		conn, err = transport.Dialer.Dial(ctx, "tcp", address)
	} else {
		conn, err = defaultDialer.DialContext(ctx, "tcp", address)
	}

	return conn, err
}

func (transport *Transport) dialProxy(ctx context.Context, conn net.Conn, scheme, host string, port uint16, creds *proxy.Creds) error {
	if scheme == "http" {
		return nil
	}
	return catch.CallWithTimeoutContextErr(ctx, transport.ProxyDialTimeout, func(ctx context.Context) error {
		var err error
		switch scheme {
		case "https":
			err = proxy.DialHttps(conn, host, port, creds)
		case "socks5", "socks5h":
			_, err = proxy.DialSocks5(conn, host, port, creds)
		default:
			panic(fmt.Sprintf("not implemented proxy '%s' scheme dialer", scheme))
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

func WithTransportHijacker(ctx context.Context) (*TransportHijacker, context.Context) {
	hijacker := &TransportHijacker{}
	return hijacker, context.WithValue(ctx, transportHijackerKey, hijacker)
}

type TransportHijacker struct {
	Conn net.Conn
}
