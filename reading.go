package giglet

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"slices"
	"strconv"

	"golang.org/x/net/http/httpguts"
)

var (
	readingOp                       = specs.GigletOp("reading")
	ErrorReadingUnsupportedEncoding = &specs.GigletError{
		Op:  readingOp,
		Err: errors.New("encoding not supported"),
	}
	ErrorReadingTimeout = &specs.GigletError{
		Op:  readingOp,
		Err: errors.New("timeout"),
	}
)

func readRequest(ctx context.Context, reader *bufio.Reader) (*httpRequest, error) {
	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	line, err := internal.ReadBufferLine(reader, HeadlineMaxLength)
	if err != nil {
		return nil, err
	}

	method, rawurl, protoMajor, protoMinor, ok := parseRequestHeadline(line)
	if !ok {
		return nil, &statusErrorResponse{
			code: specs.StatusCodeRequestURITooLong,
			text: "http: invalid headline",
		}
	}
	if protoMajor != 1 && (protoMajor != 2 || protoMinor != 0 || method != specs.HttpMethodPreface) {
		return nil, &statusErrorResponse{
			code: specs.StatusCodeNotImplemented,
			text: fmt.Sprintf("http: unsupported http version %d.%d", protoMajor, protoMinor),
		}
	}

	var url *specs.Url
	if url, err = specs.ParseUrl(rawurl); err != nil {
		return nil, &statusErrorResponse{
			code: specs.StatusCodeMisdirectedRequest,
			text: fmt.Sprintf("http: invalid request url \"%s\"", rawurl),
		}
	}

	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	req := new(httpRequest)
	req.method = method
	req.protoMajor, req.protoMinor = protoMajor, protoMinor
	req.url = url

	headers, cookies, err := parseHeaders(ctx, reader)
	if err != nil {
		return nil, err
	}

	if req.ProtoAtLeast(1, 1) { // [FEATURE]: Add chunked transfer
		if raw := headers["Transfer-Encoding"]; len(raw) > 0 { // !strings.EqualFold(raw, "chunked")
			return nil, ErrorReadingUnsupportedEncoding
		}
	}

	// RFC 7230, section 5.3: Must treat
	//	GET /index.html HTTP/1.1
	//	Host: www.google.com
	// and
	//	GET http://www.google.com/index.html HTTP/1.1
	//	Host: doesnt matter
	// the same. In the second case, any Host line is ignored.
	if host, has := headers["Host"]; has && len(host) > 0 && !httpguts.ValidHostHeader(host) {
		headers["Host"] = req.url.Host
	}

	// RFC 7234, section 5.4: Should treat
	if pragma, has := headers["Pragma"]; has && pragma == "no-cache" {
		headers["Cache-Control"] = "no-cache"
	}

	req.header = specs.NewReadOnlyHeader(headers, cookies)

	if req.method.IsPostable() {
		contentLength := req.Header().ContentLength()
		if contentLength > 0 {
			req.body = io.LimitReader(reader, contentLength)
		}
	}
	return req, nil
}

func readResponse(ctx context.Context, reader *bufio.Reader) (*HttpClientResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	line, err := internal.ReadBufferLine(reader, HeadlineMaxLength)
	if err != nil {
		return nil, err
	}

	status, protoMajor, protoMinor, ok := parseResponseHeadline(line)
	if !ok {
		return nil, &specs.GigletError{
			Op:  readingOp,
			Err: errors.New("invalid headline"),
		}
	}
	if protoMajor != 1 {
		return nil, &specs.GigletError{
			Op:  readingOp,
			Err: fmt.Errorf("unsupported http version %d.%d", protoMajor, protoMinor),
		}
	}

	headers, cookies, err := parseHeaders(ctx, reader)
	if err != nil {
		return nil, err
	}

	resp := &HttpClientResponse{
		status: status,
		header: specs.NewReadOnlyHeader(headers, cookies),
	}

	return resp, nil
}

func parseHeaders(ctx context.Context, reader *bufio.Reader) (map[string]string, map[string]*specs.Cookie, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ErrorCancelled
	default:
	}

	// The first line cannot start with a leading space.
	if buf, err := reader.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		line, err := internal.ReadBufferLine(reader, HeadlineMaxLength)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, &specs.GigletError{
			Op:  "headers/reading",
			Err: fmt.Errorf("malformed header initial line: %s", line),
		}
	}

	headers := map[string]string{}
	cookies := map[string]*specs.Cookie{}

	var key, value []byte
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ErrorCancelled
		default:
		}

		line, err := internal.ReadBufferLine(reader, 0)
		if err != nil {
			return nil, nil, &specs.GigletError{
				Op:  "headers/reading",
				Err: err,
			}
		} else if len(line) == 0 {
			return headers, cookies, nil
		}

		line = bytes.TrimLeft(line, " ")
		if len(line) < 2 {
			continue
		}
		if value != nil && len(value) != 0 && line[0] == '\t' {
			value = append(value, line[1:]...)
			continue
		}

		var ok bool
		key, value, ok = bytes.Cut(line, directColon)
		if !ok || len(key) == 0 || len(value) == 0 {
			continue
		}

		skey, sval := internal.BufferToString(internal.TitleCaseBytes(key)), string(value)
		if httpguts.ValidHeaderFieldName(skey) && httpguts.ValidHeaderFieldValue(sval) {
			if skey == "Cookie" {
				for hkey, hval := range parseCookieHeader(sval) {
					cookies[hkey] = &specs.Cookie{
						Name:  hkey,
						Value: hval,
					}
				}
			} else if skey == "Set-Cookie" {
				cookie := parseSetCookieHeader(sval)
				if cookie != nil {
					cookies[cookie.Name] = cookie
				}
			} else {
				headers[skey] = sval
			}
		}
		key, value = nil, nil
	}
}

// parse first line: GET /index.html HTTP/1.0
func parseRequestHeadline(line []byte) (method specs.HttpMethod, url string, major, minor uint16, res bool) {
	var proto []byte
	for i, b := range line {
		if b == ' ' {
			if method == "" {
				method = specs.HttpMethod(internal.BufferToString(line[:i]))
				if i < 3 {
					break
				}
			} else {
				if i-len(method) <= 1 ||
					len(line)-i != 9 {
					break
				}

				url = internal.BufferToString(line[len(method)+1 : i])
				proto = line[i+1:]
			}
		}
	}
	if method == "" || url == "" || proto == nil {
		return
	}
	major, minor, res = parseHTTPVersion(proto)
	return
}

// parse first line: HTTP/1.0 200 OK
func parseResponseHeadline(line []byte) (status specs.StatusCode, major, minor uint16, res bool) {
	if !slices.Equal(line[:5], httpVersionPrefix) || len(line) < 14 ||
		line[5] != '1' || line[6] != '.' || line[8] != ' ' || line[12] != ' ' {
		return
	}
	major, minor, res = parseHTTPVersion(line[:8])
	if !res {
		return
	}
	res = false
	code, err := strconv.ParseUint(internal.BufferToString(line[9:12]), 10, 16)
	if err != nil {
		return
	}
	status = specs.StatusCode(code)
	res = true
	return
}

func parseHTTPVersion(vers []byte) (major, minor uint16, ok bool) {
	if bytes.EqualFold(vers, httpV10) {
		return 1, 0, true
	} else if bytes.EqualFold(vers, httpV11) {
		return 1, 1, true
	} else if bytes.EqualFold(vers, httpV2) {
		return 2, 0, true
	} else if !bytes.HasPrefix(vers, httpVersionPrefix) ||
		len(vers) != 8 || vers[6] != '.' {
		return 0, 0, false
	}

	maj, err := strconv.ParseUint(internal.BufferToString(vers[5:6]), 10, 16)
	if err != nil {
		return 0, 0, false
	}
	min, err := strconv.ParseUint(internal.BufferToString(vers[7:8]), 10, 16)
	if err != nil {
		return 0, 0, false
	}
	return uint16(maj), uint16(min), true
}
