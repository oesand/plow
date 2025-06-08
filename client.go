package giglet

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"
)

func DefaultClient() *Client {
	dialer := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 10 * time.Second,
	}
	return &Client{
		ReadLineMaxLength:   1024,
		HeadMaxLength:       8 * 1024,
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		DialContext:         dialer.DialContext,
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

	// TLSClientConfig specifies the TLS configuration to use with
	// tls.Client.
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
	// If DialContext is nil (and the deprecated Dial below is also nil),
	// then the transport dials using package net.
	//
	// DialContext runs concurrently with calls to RoundTrip.
	// A RoundTrip call that initiates a dial may end up using
	// a connection dialed previously when the earlier connection
	// becomes idle before the later DialContext completes.
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

	mu sync.RWMutex
}

func (cln *Client) Make(request ClientRequest) (ClientResponse, error) {
	return cln.MakeContext(context.Background(), request)
}

func (cln *Client) MakeContext(ctx context.Context, request ClientRequest) (ClientResponse, error) {
	if cln == nil {
		return nil, validationErr("nil Client pointer")
	}
	if ctx == nil {
		return nil, validationErr("nil Context pointer")
	}
	if request == nil {
		return nil, validationErr("nil Request pointer")
	}

	_url := request.Url()
	url := &_url
	if !(url.Scheme == "http" || url.Scheme == "https") || url.Host == "" {
		return nil, validationErr("invalid request url '%s'", url)
	}

	method := request.Method()
	if !method.IsValid() {
		return nil, validationErr("invalid request method '%s'", method)
	}

	for {
		if err := catch.CatchContextCancel(ctx); err != nil {
			return nil, err
		}

		writer, _ := request.(BodyWriter)
		resp, err := cln.send(ctx, method, url, request.Header(), writer)

		if err != nil {
			return nil, catch.CatchCommonErr(err)
		}

		if err = catch.CatchContextCancel(ctx); err != nil {
			return nil, err
		}

		code := resp.StatusCode()
		if code.IsRedirect() {
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

			baseUrl := *url
			url, err = specs.ParseUrl(location)
			if err != nil {
				return nil, specs.NewOpError("redirect", "cannot parse location header url")
			}
			request.Header().Set("Host", url.Host)

			url.Scheme = baseUrl.Scheme
			if url.Host == "" {
				url.Host = baseUrl.Host
			}
			if url.Port == 0 {
				url.Port = baseUrl.Port
			}
			continue
		}

		return resp, nil
	}
}

func (cln *Client) send(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
	if !(url.Scheme == "http" || url.Scheme == "https") || url.Host == "" {
		return nil, validationErr("invalid request url '%s'", url.String())
	}

	if url.Port == 0 {
		switch url.Scheme {
		case "http", "":
			url.Port = 80
		case "https":
			url.Port = 443
		}
	}

	isTls := url.Scheme == "https"
	address := url.Host + ":" + strconv.FormatUint(uint64(url.Port), 10)
	conn, err := cln.dial(ctx, address, url.Host, isTls)

	if err != nil {
		return nil, err
	}

	if cln.Jar != nil {
		cln.mu.RLock()
		for cookie := range cln.Jar.Cookies(url.Host) {
			header.SetCookie(cookie)
		}
		cln.mu.RUnlock()
	}

	if cln.Header != nil {
		cln.mu.RLock()
		for name, value := range cln.Header.All() {
			header.Set(name, value)
		}
		for cookie := range cln.Header.Cookies() {
			header.SetCookie(cookie)
		}
		cln.mu.RUnlock()
	}

	if !header.Has("Accept") {
		header.Set("Accept", "*/*")
	}
	if !header.Has("Accept-Encoding") &&
		!header.Has("Range") &&
		method != specs.HttpMethodHead {
		header.Set("Accept-Encoding", "gzip")
	}
	if !header.Has("Host") {
		header.Set("Host", url.Host)
	}
	if url.Username != "" && !header.Has("Authorization") {
		header.Set("Authorization", "Basic "+specs.BasicAuthHeader(url.Username, url.Password))
	}

	// If protocol switcher
	if header.Get("Connection") != "close" &&
		!(header.Has("Upgrade") && strings.EqualFold(header.Get("Connection"), "Upgrade")) {
		header.Set("Connection", "close")
	}

	if cln.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(cln.WriteTimeout))
	}

	_, err = writing.WriteRequestHead(conn, method, url, header)

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
		err = writer.WriteBody(conn)
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

	if err = catch.CatchCommonErr(err); err != nil {
		conn.Close()
		if _, ok := err.(*specs.GigletError); ok {
			return nil, err
		}
		return nil, &specs.GigletError{
			Op:  "read",
			Err: err,
		}
	} else if err = catch.CatchContextCancel(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	if cln.Jar != nil {
		cln.mu.Lock()
		cln.Jar.SetCookiesIter(url.Host, resp.Header().Cookies())
		cln.mu.Unlock()
	}

	if !method.CanHaveResponseBody() || !resp.StatusCode().HaveBody() || resp.StatusCode().IsRedirect() {
		_ = conn.Close()
	} else {
		extraBuffered, _ := headerReader.Peek(headerReader.Buffered())
		reader := io.MultiReader(bytes.NewReader(extraBuffered), conn)
		chainedClosers := []io.Closer{conn}

		switch resp.Header().Get("Transfer-Encoding") {
		case "chunked":
			reader = httputil.NewChunkedReader(reader)
		}

		switch resp.Header().Get("Content-Encoding") {
		case string(specs.GzipContentEncoding):
			gzreader, err := gzip.NewReader(reader)
			if err != nil {
				return nil, &specs.GigletError{
					Op:  "encoding",
					Err: err,
				}
			}
			reader = gzreader
			chainedClosers = append(chainedClosers, gzreader)
		}

		var body io.ReadCloser
		body = stream.ReadClose(func(p []byte) (int, error) {
			if err = catch.CatchContextCancel(ctx); err != nil {
				body.Close()
				return -1, err
			}
			i, err := reader.Read(p)
			return i, catch.CatchCommonErr(err)
		}, func() error {
			for closer := range utils.ReverseIter(chainedClosers) {
				err = closer.Close()
			}
			return err
		})

		resp.SetBody(body)

		if err = catch.CatchContextCancel(ctx); err != nil {
			body.Close()
			return nil, err
		}
	}

	return resp, nil
}

func (cln *Client) dial(ctx context.Context, address, host string, isTls bool) (net.Conn, error) {
	var conn net.Conn
	var err error
	if cln.DialContext != nil {
		conn, err = cln.DialContext(ctx, "tcp", address)
	} else {
		conn, err = zeroDialer.DialContext(ctx, "tcp", address)
	}

	if err = catch.CatchCommonErr(err); err != nil {
		if _, ok := err.(*specs.GigletError); ok {
			return nil, err
		}
		return nil, &specs.GigletError{
			Op:  "dial",
			Err: err,
		}
	}

	if err = catch.CatchContextCancel(ctx); err != nil {
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
				return tlsConn, err
			})
		}

		if err = catch.CatchCommonErr(err); err != nil {
			conn.Close()
			if _, ok := err.(*specs.GigletError); ok {
				return nil, err
			}
			return nil, &specs.GigletError{
				Op:  "tls",
				Err: err,
			}
		}
		return newConn, nil
	}

	return conn, nil
}
