package giglet

import (
	"errors"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"strconv"
)

var (
	writingOp           = specs.GigletOp("writing")
	ErrorWritingTimeout = &specs.GigletError{
		Op:  writingOp,
		Err: errors.New("timeout"),
	}
)

func writeResponseHead(writer io.Writer, is11 bool, code specs.StatusCode, header *specs.Header) (int, error) {
	if !code.IsValid() {
		code = specs.StatusCodeOK
	}
	if header == nil {
		return -1, validationErr("giglet: invalid response header")
	}

	// Headline
	var buffer []byte
	if is11 {
		buffer = append(buffer, httpV11...)
	} else {
		buffer = append(buffer, httpV10...)
	}

	buffer = append(buffer, ' ')
	buffer = strconv.AppendUint(buffer, uint64(code), 10)
	buffer = append(buffer, ' ')
	buffer = append(buffer, code.Detail()...)
	buffer = append(buffer, directCrlf...)

	// Headers
	buffer = append(buffer, header.Bytes()...)
	buffer = append(buffer, header.SetCookieHeaderBytes()...)
	buffer = append(buffer, directCrlf...)
	return writer.Write(buffer)
}

func writeRequestHead(writer io.Writer, method specs.HttpMethod, url *specs.Url, header *specs.Header) (int, error) {
	if !method.IsValid() {
		return -1, validationErr("invalid request method")
	}
	if url == nil {
		return -1, validationErr("invalid request url")
	}
	if header == nil {
		return -1, validationErr("invalid request header")
	}

	// Headline
	buffer := internal.StringToBuffer(string(method))
	buffer = append(buffer, ' ')
	buffer = append(buffer, internal.StringToBuffer(url.Path)...)

	query := url.Query()
	if query != "" {
		buffer = append(buffer, '?')
		buffer = append(buffer, internal.StringToBuffer(query)...)
	}

	buffer = append(buffer, ' ')
	buffer = append(buffer, httpV11...)
	buffer = append(buffer, directCrlf...)

	// Headers
	buffer = append(buffer, header.Bytes()...)
	buffer = append(buffer, header.CookieHeaderBytes()...)
	buffer = append(buffer, directCrlf...)
	return writer.Write(buffer)
}
