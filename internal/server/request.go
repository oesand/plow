package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/parsing"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
	"net"
	"strings"
)

func ReadRequest(
	ctx context.Context, conn net.Conn, reader *bufio.Reader,
	lineLimit int64, totalLimit int64,
) (*HttpRequest, error) {
	select {
	case <-ctx.Done():
		return nil, specs.ErrCancelled
	default:
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

	if err = catch.CatchContextCancel(ctx); err != nil {
		conn.Close()
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

	if protoMajor > 1 || (protoMajor == 1 && protoMinor >= 0) { // [FEATURE]: Add chunked transfer
		if raw, has := header.TryGet("Transfer-Encoding"); has && len(raw) > 0 && !strings.EqualFold(raw, "chunked") {
			return nil, &ErrorResponse{
				Code: specs.StatusCodeNotImplemented,
				Text: "http: unsupported transfer encoding",
			}
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
		header.Set("Host", url.Host)
	}

	// RFC 7234, section 5.4: Should treat
	if pragma, has := header.TryGet("Pragma"); has && pragma == "no-cache" {
		header.Set("Cache-Control", "no-cache")
	}

	var selectedEncoding specs.ContentEncoding

	if encoding, has := header.TryGet("Accept-Encoding"); has {
		variants := strings.Split(encoding, ",")
		for _, variant := range variants {
			if strings.Contains(strings.ToLower(variant), "gzip") {
				selectedEncoding = specs.GzipContentEncoding
				break
			}
		}
		if selectedEncoding == specs.UnknownContentEncoding {
			return nil, &ErrorResponse{
				Code: specs.StatusCodeNotImplemented,
				Text: "http: has not supported encoding from accept-encoding, supported: gzip",
			}
		}
	}

	req := &HttpRequest{
		method:     method,
		protoMajor: protoMajor,
		protoMinor: protoMinor,
		url:        url,
		context:    ctx,
		header:     header,

		SelectedEncoding: selectedEncoding,
	}

	return req, nil
}
