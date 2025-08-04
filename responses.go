package giglet

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
)

// NewHeaderResponse creates new [HeaderResponse]
//
// if status code unspecified then [specs.StatusCodeOK] will be set
func NewHeaderResponse(statusCode specs.StatusCode) *HeaderResponse {
	if statusCode == specs.StatusCodeUndefined {
		statusCode = specs.StatusCodeOK
	}
	return &HeaderResponse{
		statusCode: statusCode,
		header:     specs.NewHeader(),
	}
}

// HeaderResponse is basic implementation of the [Response] without body
// to be sent by the [Server].
type HeaderResponse struct {
	_ internal.NoCopy

	statusCode specs.StatusCode
	header     *specs.Header
}

// StatusCode implementation for [Response.StatusCode] interface
//
// if status code unspecified then [specs.StatusCodeOK] will be set
func (resp *HeaderResponse) StatusCode() specs.StatusCode {
	if resp.statusCode == specs.StatusCodeUndefined {
		resp.statusCode = specs.StatusCodeOK
	}
	return resp.statusCode
}

// Header implementation for [Response.Header] interface
func (resp *HeaderResponse) Header() *specs.Header {
	if resp.header == nil {
		resp.header = specs.NewHeader()
	}
	return resp.header
}

// EmptyResponse is implementation for the [Response] without body
// to be sent by the [Server].
//
// if status code unspecified then [specs.StatusCodeOK] will be set
func EmptyResponse(statusCode specs.StatusCode, configure ...func(Response)) Response {
	resp := NewHeaderResponse(statusCode)

	for _, conf := range configure {
		conf(resp)
	}

	return resp
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
	if buffer == nil {
		panic("giglet/response: passed nil buffer")
	}

	resp := &bufferResponse{
		HeaderResponse: *NewHeaderResponse(statusCode),
		buffer:         buffer,
		contentLength:  int64(len(buffer)),
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Type", contentType)

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

type bufferResponse struct {
	HeaderResponse
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
		panic("giglet/response: passed nil stream")
	}

	resp := &streamResponse{
		HeaderResponse: *NewHeaderResponse(statusCode),
		stream:         stream,
		contentLength:  contentLength,
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Type", contentType)

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

type streamResponse struct {
	HeaderResponse
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
