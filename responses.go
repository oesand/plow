package giglet

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"strconv"
)

func NewEmptyResponse(contentType specs.ContentType, configure ...func(response Response)) Response {
	resp := &HeaderResponse{}

	if contentType != specs.ContentTypeUndefined {
		resp.Header().Set("Content-Type", string(contentType))
	}

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

func NewBufferResponse(buffer []byte, contentType specs.ContentType, configure ...func(response Response)) Response {
	resp := &bufferResponse{buffer: buffer}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Length", strconv.Itoa(len(buffer)))
	resp.Header().Set("Content-Type", string(contentType))

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

func NewTextResponse(text string, contentType specs.ContentType, configure ...func(response Response)) Response {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return NewBufferResponse(internal.StringToBuffer(text), contentType, configure...)
}

type bufferResponse struct {
	HeaderResponse
	buffer []byte
}

func (resp *bufferResponse) WriteBody(writer io.Writer) error {
	_, err := writer.Write(resp.buffer)
	return err
}

func NewStreamResponse(stream io.Reader, size uint64, contentType specs.ContentType, configure ...func(response Response)) Response {
	resp := &streamResponse{stream: stream}
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	if size > 0 {
		resp.Header().Set("Content-Length", strconv.FormatUint(size, 10))
	}
	resp.Header().Set("Content-Type", string(contentType))

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

type streamResponse struct {
	HeaderResponse
	stream io.Reader
}

func (resp *streamResponse) WriteBody(writer io.Writer) error {
	_, err := io.Copy(writer, resp.stream)
	return err
}
