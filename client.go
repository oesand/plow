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
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"
)

func DefaultClient() *Client {
	return &Client{
		ReadLineMaxLength:   1024,
		HeadMaxLength:       8 * 1024,
		MaxBodySize:         10 << 20, // 10 mb
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxRedirectCount:    DefaultMaxRedirectCount,
	}
}

func MakeRequest(request ClientRequest) (ClientResponse, error) {
	return DefaultClient().Make(request)
}

type Client struct {
	// ReadLineMaxLength maximum size in bytes
	// to read lines in the response
	// such as headers and headlines
	// If zero there is no limit
	ReadLineMaxLength int64

	// HeadMaxLength maximum size in bytes
	// to read headline and headers together
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

	// MaxRedirectCount maximum number of redirects
	// before getting an error.
	// if not specified is used DefaultMaxRedirectCount
	MaxRedirectCount int

	// ReadTimeout is the maximum duration for server the entire
	// response, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the request. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	// Header specifies independent request header and cookies
	//
	// The Header is used to insert headers and cookies
	// into every outbound Request independent of url.
	// The Header is consulted for every redirect that the Client follows.
	//
	// If Header is nil, headers and cookies are only sent
	// if they are explicitly set on the Request.
	Header *specs.Header

	// Jar specifies the cookie jar with dependent to url
	//
	// The Jar is used to insert relevant requested url cookies
	// into every outbound Request and is updated
	// with the cookie values of every inbound Response.
	// The Jar is consulted for every redirect that the Client follows.
	//
	// If Jar is nil, cookies are only sent
	// if they are explicitly set on the Request.
	Jar *specs.CookieJar

	// TLSClientConfig specifies the TLS configuration
	// to use with tls.Client.
	// If nil, the default configuration is used.
	// If non-nil, HTTP/2 support may not be enabled by default.
	TLSConfig *tls.Config

	// TLSHandshakeContext specifies an optional dial function for creating
	// TLS connections and gone handshake for non-proxied HTTPS requests.
	//
	// If TLSHandshakeContext is nil will be used tls.Client.
	//
	// If TLSHandshakeContext is set, it is assumed
	// that the returned net.Conn has already gone through the TLS handshake.
	TLSHandshakeContext func(ctx context.Context, conn net.Conn, host string) (net.Conn, error)

	// DialContext specifies the dial function for creating unencrypted TCP connections.
	// If DialContext is nil then the transport dials using package net.
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	mu sync.RWMutex
}

func (cln *Client) Make(request ClientRequest) (ClientResponse, error) {
	if request == nil {
		panic("nil request pointer")
	}
	return cln.MakeContext(context.Background(), request)
}

func (cln *Client) MakeContext(ctx context.Context, request ClientRequest) (ClientResponse, error) {
	if cln == nil {
		panic("nil client pointer")
	}
	if ctx == nil {
		panic("nil context pointer")
	}
	if request == nil {
		panic("nil request pointer")
	}

	url := request.Url()
	if url.Scheme == "" {
		url.Scheme = "https"
	}

	if !(url.Scheme == "http" || url.Scheme == "https") {
		panic(fmt.Sprintf("invalid request url '%s' scheme", url.Scheme))
	}

	if url.Host == "" {
		panic("empty request url host")
	}

	if url.Port == 0 {
		switch url.Scheme {
		case "http":
			url.Port = 80
		case "https":
			url.Port = 443
		}
	}

	method := request.Method()
	if !method.IsValid() {
		panic(fmt.Sprintf("invalid request method '%s'", method))
	}

	maxRedirectCount := DefaultMaxRedirectCount
	if cln.MaxRedirectCount > 0 {
		maxRedirectCount = cln.MaxRedirectCount
	}

	var redirectCount int
	for {
		if err := catch.CatchContextCancel(ctx); err != nil {
			return nil, err
		}

		resp, err := cln.send(ctx, method, url, request)

		if err != nil {
			return nil, catch.CatchCommonErr(err)
		}

		if err = catch.CatchContextCancel(ctx); err != nil {
			return nil, err
		}

		if resp.Hijacked {
			return resp, nil
		}

		code := resp.StatusCode()
		if code.IsRedirect() {
			if redirectCount >= maxRedirectCount {
				return nil, specs.NewOpError("redirect", "too many redirects")
			}
			redirectCount++

			if (code == specs.StatusCodeMovedPermanently ||
				code == specs.StatusCodeSeeOther ||
				code == specs.StatusCodeFound) &&
				(method != specs.HttpMethodGet &&
					method != specs.HttpMethodHead) {
				method = specs.HttpMethodGet
			}

			location := resp.Header().Get("Location")
			if location == "" {
				return nil, specs.NewOpError("redirect", "empty Location header")
			}

			var redirectUrl *specs.Url
			redirectUrl, err = specs.ParseUrl(location)
			if err != nil {
				return nil, specs.NewOpError("redirect", "cannot parse location header url")
			}

			if !(redirectUrl.Scheme == "" || redirectUrl.Scheme == "http" || redirectUrl.Scheme == "https") {
				return nil, specs.NewOpError("redirect", "invalid request url '%s' scheme", url.Scheme)
			}

			redirectUrl.Scheme = url.Scheme
			if redirectUrl.Host == "" {
				redirectUrl.Host = url.Host
				redirectUrl.Port = url.Port
			}
			request.Header().Set("Host", redirectUrl.Host)
			url = *redirectUrl

			continue
		}

		return resp, nil
	}
}

func (cln *Client) send(ctx context.Context, method specs.HttpMethod, url specs.Url, request ClientRequest) (*client.HttpClientResponse, error) {
	header := request.Header()
	if header == nil {
		panic("nil request.header pointer")
	}

	writer, _ := request.(BodyWriter)
	hijacker, _ := request.(HijackRequest)

	if cln.Jar != nil {
		cln.mu.RLock()
		for cookie := range cln.Jar.Cookies(url.Host) {
			if !header.HasCookie(cookie.Name) {
				header.SetCookie(cookie)
			}
		}
		cln.mu.RUnlock()
	}

	if cln.Header != nil {
		cln.mu.RLock()
		for name, value := range cln.Header.All() {
			if !header.Has(name) {
				header.Set(name, value)
			}
		}
		for cookie := range cln.Header.Cookies() {
			if !header.HasCookie(cookie.Name) {
				header.SetCookie(cookie)
			}
		}
		cln.mu.RUnlock()
	}

	if !header.Has("Accept") {
		header.Set("Accept", "*/*")
	}
	if !header.Has("Accept-Encoding") &&
		!header.Has("Range") &&
		method != specs.HttpMethodHead {
		header.Set("Accept-Encoding", encoding.DefaultAcceptEncoding)
	}
	if !header.Has("Host") {
		host, err := httpguts.PunycodeHostPort(url.Host)
		if err != nil {
			return nil, err
		}
		header.Set("Host", host)
	}
	if url.Username != "" && !header.Has("Authorization") {
		header.Set("Authorization", "Basic "+specs.BasicAuthHeader(url.Username, url.Password))
	}

	if hijacker == nil && !header.Has("Connection") {
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

	isTls := url.Scheme == "https"
	conn, err := cln.dial(ctx, url.Host, url.Port, isTls)
	if err != nil {
		if _, ok := err.(*specs.GigletError); ok {
			return nil, err
		}
		return nil, &specs.GigletError{
			Op:  "dial",
			Err: err,
		}
	}

	if cln.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(cln.WriteTimeout))
	}

	_, err = client.WriteRequestHead(conn, method, &url, header)

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

	if cln.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Time{})
	}

	if cln.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(cln.ReadTimeout))
	}

	headerReader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(headerReader)

	resp, err := client.ReadResponse(ctx, headerReader, cln.ReadLineMaxLength, cln.HeadMaxLength)

	err = catch.CatchCommonErr(err)
	if err == nil {
		err = catch.CatchContextCancel(ctx)
	}
	if err != nil {
		conn.Close()
		return nil, err
	}

	if cln.Jar != nil {
		cln.mu.Lock()
		cln.Jar.SetCookiesIter(url.Host, resp.Header().Cookies())
		cln.mu.Unlock()
	}

	if !method.IsReplyable() || !resp.StatusCode().IsReplyable() {
		if hijacker != nil && header.Get("Connection") != "close" {
			if cln.ReadTimeout > 0 {
				conn.SetReadDeadline(time.Time{})
			}
			resp.Hijacked = true
			hijacker.Hijack(conn)
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
			if cln.MaxBodySize > 0 {
				reader = io.LimitReader(reader, cln.MaxBodySize)
			}
		} else {
			var contentLength int64
			if contentLengthString := resp.Header().Get("Content-Length"); contentLengthString != "" {
				contentLength, err = strconv.ParseInt(contentLengthString, 10, 64)
				if err == nil {
					if contentLength <= 0 {
						reader = nil
						conn.Close()
					} else if cln.MaxBodySize > 0 && contentLength >= cln.MaxBodySize {
						return nil, specs.ErrTooLarge
					}
				}
			}

			if reader != nil {
				if contentLength <= 0 && cln.MaxBodySize > 0 {
					contentLength = cln.MaxBodySize
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

		if hijacker != nil && header.Get("Connection") != "close" {
			if cln.ReadTimeout > 0 {
				conn.SetReadDeadline(time.Time{})
			}
			resp.Hijacked = true
			hijacker.Hijack(conn)
		}
	}

	return resp, nil
}

func (cln *Client) dial(ctx context.Context, host string, port uint16, isTls bool) (net.Conn, error) {
	address := host + ":" + strconv.FormatUint(uint64(port), 10)

	var conn net.Conn
	var err error
	if cln.DialContext != nil {
		conn, err = cln.DialContext(ctx, "tcp", address)
	} else {
		conn, err = defaultDialer.DialContext(ctx, "tcp", address)
	}

	if err = catch.CatchCommonErr(err); err != nil {
		return nil, err
	}

	if err = catch.CatchContextCancel(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	if isTls {
		var newConn net.Conn
		if cln.TLSHandshakeContext != nil {
			newConn, err = catch.CallWithTimeoutContext(ctx, cln.TLSHandshakeTimeout, func(ctx context.Context) (net.Conn, error) {
				return cln.TLSHandshakeContext(ctx, conn, host)
			})
		} else {
			newConn, err = catch.CallWithTimeoutContext(ctx, cln.TLSHandshakeTimeout, func(ctx context.Context) (net.Conn, error) {
				var tlsCfg *tls.Config

				if cln.TLSConfig == nil {
					tlsCfg = &tls.Config{}
				} else {
					tlsCfg = cln.TLSConfig.Clone()
				}

				if tlsCfg.ServerName == "" {
					tlsCfg.ServerName = host
				}

				tlsConn := tls.Client(conn, tlsCfg)
				err = tlsConn.HandshakeContext(ctx)
				if err != nil {
					return nil, err
				}
				return tlsConn, nil
			})
		}

		if err = catch.CatchCommonErr(err); err != nil {
			conn.Close()
			return nil, err
		}
		return newConn, nil
	}

	return conn, nil
}
