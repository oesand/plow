package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal/parsing"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	ResponseErrUnsupportedEncoding = &ErrorResponse{
		Code: specs.StatusCodeUnprocessableEntity,
		Text: "unsupported transfer encoding",
	}
	ErrorCancelled = &specs.GigletError{Err: errors.New("cancelled")}
)

func ReadRequest(
	ctx context.Context, conn net.Conn, reader *bufio.Reader, timeout time.Duration,
	lineLimit int64, totalLimit int64,
) (*HttpRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}

	line, err := utils.ReadBufferLine(reader, lineLimit)
	if err != nil {
		return nil, err
	}

	method, rawurl, protoMajor, protoMinor, ok := parsing.ParseClientRequestHeadline(line)
	if !ok {
		return nil, &ErrorResponse{
			Code: specs.StatusCodeRequestURITooLong,
			Text: "http: invalid headline",
		}
	}
	if protoMajor != 1 && (protoMajor != 2 || protoMinor != 0 || method != specs.MethodPreface) {
		return nil, &ErrorResponse{
			Code: specs.StatusCodeUnprocessableEntity,
			Text: fmt.Sprintf("http: unsupported http version %d.%d", protoMajor, protoMinor),
		}
	}

	if !method.IsValid() {
		return nil, &ErrorResponse{
			Code: specs.StatusCodeMethodNotAllowed,
			Text: fmt.Sprintf("http: unknown http method %s", method),
		}
	}

	var url *specs.Url
	if url, err = specs.ParseUrl(rawurl); err != nil {
		return nil, &ErrorResponse{
			Code: specs.StatusCodeMisdirectedRequest,
			Text: fmt.Sprintf("http: invalid request url \"%s\"", rawurl),
		}
	}

	if timeout > 0 {
		conn.SetReadDeadline(time.Time{})
	}

	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	req := &HttpRequest{
		method:     method,
		protoMajor: protoMajor,
		protoMinor: protoMinor,
		url:        url,
		context:    ctx,

		readTimeout: timeout,
	}

	header, err := parsing.ParseHeaders(ctx, reader, lineLimit, totalLimit)
	if err != nil {
		return nil, err
	}

	req.header = header

	if protoMajor > 1 || (protoMajor == 1 && protoMinor >= 0) { // [FEATURE]: Add chunked transfer
		if raw, has := header.TryGet("Transfer-Encoding"); has && len(raw) > 0 && !strings.EqualFold(raw, "chunked") {
			return nil, ResponseErrUnsupportedEncoding
		}
	}

	// RFC 7230, section 5.3: Must treat
	//	GET /index.html HTTP/1.1
	//	Host: www.google.com
	// and
	//	GET http://www.google.com/index.html HTTP/1.1
	//	Host: doesnt matter
	// the same. In the second case, any Host line is ignored.
	if host, has := header.TryGet("Host"); has && len(host) > 0 && !httpguts.ValidHostHeader(host) {
		header.Set("Host", req.url.Host)
	}

	// RFC 7234, section 5.4: Should treat
	if pragma, has := header.TryGet("Pragma"); has && pragma == "no-cache" {
		header.Set("Cache-Control", "no-cache")
	}

	if req.method.IsPostable() {
		if raw, has := header.TryGet("Content-Length"); has && len(raw) > 0 {
			if contentLength, err := strconv.ParseInt(raw, 10, 64); err != nil {
				req.body = io.LimitReader(reader, contentLength)
			}
		}
	}

	return req, nil
}
