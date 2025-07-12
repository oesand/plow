package giglet

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
)

func NewEmptyResponse(statusCode specs.StatusCode, configure ...func(Response)) Response {
	resp := newEmptyResponse(statusCode)

	for _, conf := range configure {
		conf(resp)
	}

	return resp
}

func newEmptyResponse(statusCode specs.StatusCode) *HeaderResponse {
	if statusCode == specs.StatusCodeUndefined {
		statusCode = specs.StatusCodeOK
	}
	return &HeaderResponse{
		statusCode: statusCode,
		header:     specs.NewHeader(),
	}
}

func NewRedirectResponse(url string, configure ...func(Response)) Response {
	resp := NewEmptyResponse(specs.StatusCodeTemporaryRedirect, configure...)
	resp.Header().Set("Location", url)
	return resp
}

func NewPermanentRedirectResponse(url string, configure ...func(Response)) Response {
	resp := NewEmptyResponse(specs.StatusCodePermanentRedirect, configure...)
	resp.Header().Set("Location", url)
	return resp
}

type HeaderResponse struct {
	_ internal.NoCopy

	statusCode specs.StatusCode
	header     *specs.Header
}

func (resp *HeaderResponse) StatusCode() specs.StatusCode {
	if resp.statusCode == specs.StatusCodeUndefined {
		resp.statusCode = specs.StatusCodeOK
	}
	return resp.statusCode
}

func (resp *HeaderResponse) Header() *specs.Header {
	if resp.header == nil {
		resp.header = specs.NewHeader()
	}
	return resp.header
}

func NewTextResponse(text string, contentType string, statusCode specs.StatusCode, configure ...func(Response)) Response {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return NewBufferResponse([]byte(text), contentType, statusCode, configure...)
}

func NewBufferResponse(buffer []byte, contentType string, statusCode specs.StatusCode, configure ...func(Response)) Response {
	resp := &bufferResponse{
		HeaderResponse: *newEmptyResponse(statusCode),
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

func NewStreamResponse(stream io.Reader, contentType string, contentLength int64, statusCode specs.StatusCode, configure ...func(Response)) Response {
	if stream == nil {
		panic("giglet/response: passed nil stream")
	}
	if contentLength < 0 {
		panic("giglet/response: invalid content length")
	}

	resp := &streamResponse{
		HeaderResponse: *newEmptyResponse(statusCode),
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
