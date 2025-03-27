package giglet

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

func MakeRequest(request *HttpClientRequest) (*HttpClientResponse, error) {
	cln := Client{}
	return cln.Make(request)
}

type Client struct {
	// ReadTimeout is the maximum duration for reading the entire
	// response, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the request. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

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

	// DialTLSHandshakeContext specifies an optional dial function for creating
	// TLS connections and gone handshake for non-proxied HTTPS requests.
	//
	// If DialTLSHandshakeContext is nil will be used tls.Client.
	//
	// If DialTLSHandshakeContext is set, it is assumed
	// that the returned net.Conn has already gone through the TLS handshake.
	DialTLSHandshakeContext func(ctx context.Context, conn net.Conn, host string) (net.Conn, error)
}

func (client *Client) Make(request *HttpClientRequest) (*HttpClientResponse, error) {
	return client.MakeContext(context.Background(), request)
}

func (client *Client) applyReadTimeout(conn net.Conn) {
	if client.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(client.ReadTimeout))
	}
}

func (client *Client) applyWriteTimeout(conn net.Conn) {
	if client.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(client.WriteTimeout))
	}
}

func (client *Client) MakeContext(ctx context.Context, request *HttpClientRequest) (*HttpClientResponse, error) {
	if ctx == nil {
		return nil, validationErr("nil Context pointer")
	}
	if client == nil {
		return nil, validationErr("nil Client pointer")
	}
	if request == nil {
		return nil, validationErr("nil Request pointer")
	}

	_url := request.Url()
	url := &_url

	method := request.Method()

	if !(url.Scheme == "http" || url.Scheme == "https") || url.Host == "" {
		return nil, validationErr("invalid request url '%s'", method)
	}
	if !method.IsValid() {
		return nil, validationErr("invalid request method")
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ErrorCancelled
		default:
		}

		resp, err := client.send(ctx, method, url, request.Header(), request)

		if err != nil {
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
				return resp, nil
			}
			baseUrl := *url
			url, err = specs.ParseUrl(location)
			if err != nil {
				return nil, &specs.GigletError{
					Op:  "serve/redirect",
					Err: err,
				}
			}
			request.Header().Set("Host", url.Host)
			if url.Scheme == "" {
				url.Scheme = baseUrl.Scheme
			}
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

func (client *Client) send(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (*HttpClientResponse, error) {
	if !(url.Scheme == "http" || url.Scheme == "https") || url.Host == "" {
		return nil, validationErr("invalid request url '%s'", string(method))
	}

	if url.Port == 0 {
		switch url.Scheme {
		case "http", "":
			url.Port = 80
		case "https":
			url.Port = 443
		}
	}

	address := url.Host + ":" + strconv.FormatUint(uint64(url.Port), 10)
	conn, err := zeroDialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, &specs.GigletError{
			Op:  "dialing",
			Err: err,
		}
	}

	if url.Scheme == "https" {
		if client.DialTLSHandshakeContext != nil {
			conn, err = client.DialTLSHandshakeContext(ctx, conn, url.Host)
		} else {
			var tlsCfg *tls.Config

			if client.TLSConfig == nil {
				tlsCfg = &tls.Config{}
			} else {
				tlsCfg = client.TLSConfig.Clone()
			}

			if tlsCfg.ServerName == "" {
				tlsCfg.ServerName = url.Host
			}
			tlsConn := tls.Client(conn, tlsCfg)
			err = tlsConn.HandshakeContext(ctx)
			conn = tlsConn
		}

		if err != nil {
			return nil, &specs.GigletError{
				Op:  "tls",
				Err: err,
			}
		}
	}

	if client.Jar != nil {
		for cookie := range client.Jar.Cookies(url) {
			header.SetCookie(cookie)
		}
	}
	if client.Header != nil {
		for name, value := range client.Header.All() {
			header.Set(name, value)
		}
		for cookie := range client.Header.Cookies() {
			header.SetCookie(cookie)
		}
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
		!(header.Has("Upgrade") &&
			httpguts.HeaderValuesContainsToken(
				strings.Split(header.Get("Connection"), ", "), "Upgrade")) {
		header.Set("Connection", "close")
	}

	client.applyWriteTimeout(conn)
	_, err = writeRequestHead(conn, method, url, header)

	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return nil, ErrorWritingTimeout
		}
		return nil, err
	}

	if method.IsPostable() && writer != nil {
		err = writer.WriteBody(conn)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				return nil, ErrorWritingTimeout
			}
			return nil, err
		}
	}
	conn.SetWriteDeadline(zeroTime)

	client.applyReadTimeout(conn)
	headerReader := bufioReaderPool.Get(conn)
	resp, err := readResponse(ctx, headerReader)
	extraBuffered, _ := headerReader.Peek(headerReader.Buffered())
	bufioReaderPool.Put(headerReader)
	conn.SetReadDeadline(zeroTime)

	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return nil, ErrorReadingTimeout
		}
		return nil, err
	}

	if client.Jar != nil {
		client.Jar.SetCookiesIter(url, resp.header.Cookies())
	}

	if !method.CanHaveResponseBody() || !resp.StatusCode().HaveBody() || resp.StatusCode().IsRedirect() {
		_ = conn.Close()
	} else {
		reader := io.MultiReader(bytes.NewReader(extraBuffered), conn)
		chainedClosers := []io.Closer{conn}

		switch resp.Header().Get("Transfer-Encoding") {
		case "chunked":
			chunkedBuf := bufioReaderPool.Get(reader)
			reader = httputil.NewChunkedReader(chunkedBuf)
			chainedClosers = append(chainedClosers, internal.Closer(func() error {
				bufioReaderPool.Put(chunkedBuf)
				return nil
			}))
		}

		switch resp.Header().Get("Content-Encoding") {
		case "gzip":
			gzreader, err := gzip.NewReader(reader)
			if err != nil {
				return nil, &specs.GigletError{
					Op:  readingOp,
					Err: fmt.Errorf("gzip: %s", err),
				}
			}
			reader = gzreader
			chainedClosers = append(chainedClosers, gzreader)
		}

		resp.body = internal.ReadClose(func(p []byte) (int, error) {
			client.applyReadTimeout(conn)
			return reader.Read(p)
		}, func() error {
			for closer := range internal.ReverseIter(chainedClosers) {
				err = closer.Close()
			}
			return err
		})
	}

	return resp, nil
}
