package giglet

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
)

func MakeRequest(request ClientRequest) (ClientResponse, error) {
	cln := Client{}
	return cln.Make(request)
}

func DefaultClient() *Client {
	return &Client{
		ReadLineMaxLength: 256,      // 1 MB
		HeadMaxLength:     5 * 1024, // 5 MB
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
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

func (cln *Client) Make(request ClientRequest) (ClientResponse, error) {
	return cln.MakeContext(context.Background(), request)
}

func (cln *Client) applyReadTimeout(conn net.Conn) {
	if cln.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(cln.ReadTimeout))
	}
}

func (cln *Client) applyWriteTimeout(conn net.Conn) {
	if cln.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(cln.WriteTimeout))
	}
}

func (cln *Client) MakeContext(ctx context.Context, request ClientRequest) (ClientResponse, error) {
	if ctx == nil {
		return nil, validationErr("nil Context pointer")
	}
	if cln == nil {
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

		writer, _ := request.(BodyWriter)
		resp, err := cln.send(ctx, method, url, request.Header(), writer)

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

func (cln *Client) send(ctx context.Context, method specs.HttpMethod, url *specs.Url, header *specs.Header, writer BodyWriter) (ClientResponse, error) {
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
		if cln.DialTLSHandshakeContext != nil {
			conn, err = cln.DialTLSHandshakeContext(ctx, conn, url.Host)
		} else {
			var tlsCfg *tls.Config

			if cln.TLSConfig == nil {
				tlsCfg = &tls.Config{}
			} else {
				tlsCfg = cln.TLSConfig.Clone()
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

	if cln.Jar != nil {
		for cookie := range cln.Jar.Cookies(url) {
			header.SetCookie(cookie)
		}
	}
	if cln.Header != nil {
		for name, value := range cln.Header.All() {
			header.Set(name, value)
		}
		for cookie := range cln.Header.Cookies() {
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

	cln.applyWriteTimeout(conn)
	_, err = writing.WriteRequestHead(conn, method, url, header)

	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return nil, specs.NewOpError("write", "request head write timeout")
		}
		return nil, err
	}

	if method.IsPostable() && writer != nil {
		err = writer.WriteBody(conn)
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				return nil, specs.NewOpError("write", "request body write timeout")
			}
			return nil, err
		}
	}
	conn.SetWriteDeadline(zeroTime)

	cln.applyReadTimeout(conn)
	headerReader := bufioReaderPool.Get(conn)
	resp, err := client.ReadResponse(ctx, headerReader, cln.ReadLineMaxLength, cln.HeadMaxLength)
	extraBuffered, _ := headerReader.Peek(headerReader.Buffered())
	bufioReaderPool.Put(headerReader)
	conn.SetReadDeadline(zeroTime)

	if err != nil {
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			return nil, specs.NewOpError("client/response", "timeout")
		}
		return nil, err
	}

	if cln.Jar != nil {
		cln.Jar.SetCookiesIter(url, resp.Header().Cookies())
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
			chainedClosers = append(chainedClosers, utils.Closer(func() error {
				bufioReaderPool.Put(chunkedBuf)
				return nil
			}))
		}

		switch resp.Header().Get("Content-Encoding") {
		case "gzip":
			gzreader, err := gzip.NewReader(reader)
			if err != nil {
				return nil, &specs.GigletError{
					Op:  "read/encoding/gzip",
					Err: fmt.Errorf("gzip: %s", err),
				}
			}
			reader = gzreader
			chainedClosers = append(chainedClosers, gzreader)
		}

		resp.SetBody(
			utils.ReadClose(func(p []byte) (int, error) {
				cln.applyReadTimeout(conn)
				return reader.Read(p)
			}, func() error {
				for closer := range utils.ReverseIter(chainedClosers) {
					err = closer.Close()
				}
				return err
			}))
	}

	return resp, nil
}
