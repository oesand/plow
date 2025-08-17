package plow

import (
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/specs"
	"io"
)

// EmptyResponse is implementation for the [Response] without body
// to be sent by the [Server].
//
// if status code unspecified then [specs.StatusCodeOK] will be set
func EmptyResponse(statusCode specs.StatusCode, configure ...func(Response)) Response {
	if statusCode == specs.StatusCodeUndefined {
		statusCode = specs.StatusCodeOK
	}

	resp := &emptyResponse{
		statusCode: statusCode,
		header:     specs.NewHeader(),
	}

	for _, conf := range configure {
		conf(resp)
	}

	return resp
}

type emptyResponse struct {
	_ internal.NoCopy

	statusCode specs.StatusCode
	header     *specs.Header
}

func (resp *emptyResponse) StatusCode() specs.StatusCode {
	if resp.statusCode == specs.StatusCodeUndefined {
		resp.statusCode = specs.StatusCodeOK
	}
	return resp.statusCode
}

func (resp *emptyResponse) Header() *specs.Header {
	if resp.header == nil {
		resp.header = specs.NewHeader()
	}
	return resp.header
}

// RedirectResponse is implementation for the [Response] sent by the [Server],
// based on [EmptyResponse] with [specs.StatusCodeTemporaryRedirect]
// and "Location" header provided by url.
func RedirectResponse(url string, configure ...func(Response)) Response {
	resp := EmptyResponse(specs.StatusCodeTemporaryRedirect, configure...)
	resp.Header().Set("Location", url)
	return resp
}

// PermanentRedirectResponse is implementation for the [Response] sent by the [Server],
// based on [EmptyResponse] with [specs.StatusCodePermanentRedirect]
// and "Location" header provided by url.
func PermanentRedirectResponse(url string, configure ...func(Response)) Response {
	resp := EmptyResponse(specs.StatusCodePermanentRedirect, configure...)
	resp.Header().Set("Location", url)
	return resp
}

// TextResponse is implementation for the [Response]
// with string as response body to be sent by the [Server].
//
// Content type applies as "Content-Type" header value
//
// if status code unspecified then [specs.StatusCodeOK] will be set
// if content type unspecified then [specs.ContentTypePlain] will be set
func TextResponse(statusCode specs.StatusCode, contentType string, text string, configure ...func(Response)) Response {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return BufferResponse(statusCode, contentType, []byte(text), configure...)
}

// BufferResponse is implementation for the [Response]
// with []byte as response body to be sent by the [Server].
//
// Content type applies as "Content-Type" header value
//
// if status code unspecified then [specs.StatusCodeOK] will be set
// if content type unspecified then [specs.ContentTypeRaw] will be set
func BufferResponse(statusCode specs.StatusCode, contentType string, buffer []byte, configure ...func(Response)) Response {
	return newBufferResponse(statusCode, contentType, buffer, configure...)
}

func newBufferResponse(statusCode specs.StatusCode, contentType string, buffer []byte, configure ...func(Response)) *bufferResponse {
	if buffer == nil {
		panic("passed nil buffer")
	}

	resp := &bufferResponse{
		Response:      EmptyResponse(statusCode, configure...),
		buffer:        buffer,
		contentLength: int64(len(buffer)),
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Type", contentType)

	return resp
}

type bufferResponse struct {
	Response
	buffer        []byte
	contentLength int64
}

func (resp *bufferResponse) WriteBody(writer io.Writer) error {
	_, err := writer.Write(resp.buffer)
	return err
}

func (resp *bufferResponse) ContentLength() int64 {
	return resp.contentLength
}

// StreamResponse is implementation for the [Response] that
// copy response body from [io.Reader] to be sent by the [Server].
//
// Content type applies as "Content-Type" header value
//
// if status code unspecified then [specs.StatusCodeOK] will be set
// if content type unspecified then [specs.ContentTypeRaw] will be set
func StreamResponse(statusCode specs.StatusCode, contentType string, stream io.Reader, contentLength int64, configure ...func(Response)) Response {
	if stream == nil {
		panic("passed nil stream")
	}

	resp := &streamResponse{
		Response:      EmptyResponse(statusCode, configure...),
		stream:        stream,
		contentLength: contentLength,
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Type", contentType)

	return resp
}

type streamResponse struct {
	Response
	stream        io.Reader
	contentLength int64
}

func (resp *streamResponse) WriteBody(writer io.Writer) error {
	_, err := io.Copy(writer, resp.stream)
	return err
}

func (resp *streamResponse) ContentLength() int64 {
	return resp.contentLength
}
