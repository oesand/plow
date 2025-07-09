package giglet

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
)

func NewEmptyResponse(configure ...func(response Response)) Response {
	resp := &HeaderResponse{}

	for _, conf := range configure {
		conf(resp)
	}

	return resp
}

func NewRedirectResponse(url string) Response {
	resp := &HeaderResponse{}
	resp.SetStatusCode(specs.StatusCodeTemporaryRedirect)
	resp.Header().Set("Location", url)
	return resp
}

func NewPermanentRedirectResponse(url string) Response {
	resp := &HeaderResponse{}
	resp.SetStatusCode(specs.StatusCodePermanentRedirect)
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

func (resp *HeaderResponse) SetStatusCode(code specs.StatusCode) {
	resp.statusCode = code
}

func (resp *HeaderResponse) Header() *specs.Header {
	if resp.header == nil {
		resp.header = &specs.Header{}
	}
	return resp.header
}

func NewTextResponse(text string, contentType string, configure ...func(response Response)) Response {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return NewBufferResponse([]byte(text), contentType, configure...)
}

func NewBufferResponse(buffer []byte, contentType string, configure ...func(response Response)) Response {
	resp := &bufferResponse{
		buffer:        buffer,
		contentLength: int64(len(buffer)),
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

func NewStreamResponse(stream io.Reader, contentType string, contentLength int64, configure ...func(response Response)) Response {
	if stream == nil {
		panic("giglet/response: passed nil stream")
	}
	if contentLength < 0 {
		panic("giglet/response: invalid content length")
	}

	resp := &streamResponse{
		stream:        stream,
		contentLength: contentLength,
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
