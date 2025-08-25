package server_ops

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/oesand/plow/internal/parsing"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"golang.org/x/net/http/httpguts"
	"net"
)

func ReadRequest(
	ctx context.Context, remoteAddr net.Addr,
	reader *bufio.Reader, lineLimit int64, totalLimit int64,
) (*HttpRequest, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	line, err := stream.ReadBufferLine(reader, lineLimit)
	if err != nil {
		if errors.Is(err, specs.ErrTooLarge) {
			return nil, &ErrorResponse{
				Code: specs.StatusCodeRequestHeaderFieldsTooLarge,
				Text: "http: too large header",
			}
		}
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

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	header, err := parsing.ParseHeaders(ctx, reader, lineLimit, totalLimit)
	if err != nil {
		if errors.Is(err, specs.ErrTooLarge) {
			return nil, &ErrorResponse{
				Code: specs.StatusCodeRequestHeaderFieldsTooLarge,
				Text: "http: too large header",
			}
		}
		return nil, err
	}

	// RFC 7230, section 5.3: Must treat
	//	GET /index.html HTTP/1.1
	//	Host: www.google.com
	// and
	//	GET http://www.google.com/index.html HTTP/1.1
	//	Host: doesnt matter
	// the same. In the second case, any Host line is ignored.
	if host, has := header.TryGet("Host"); has && len(host) > 0 && !httpguts.ValidHostHeader(host) {
		header.Set("Host", url.Host)
	}

	// RFC 7234, section 5.4: Should treat
	if pragma, has := header.TryGet("Pragma"); has && pragma == "no-cache" {
		header.Set("Cache-Control", "no-cache")
	}

	req := &HttpRequest{
		method:     method,
		protoMajor: protoMajor,
		protoMinor: protoMinor,
		remoteAddr: remoteAddr,
		url:        url,
		header:     header,
	}

	return req, nil
}
