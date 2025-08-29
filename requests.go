package plow

import (
	"fmt"
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/specs"
	"io"
)

// EmptyRequest is implementation for the [ClientRequest] without body
// to be sent by the [Client].
//
// if method unspecified then [specs.HttpMethodGet] will be set
func EmptyRequest(method specs.HttpMethod, url *specs.Url) ClientRequest {
	if method == "" {
		method = specs.HttpMethodGet
	}
	if !method.IsValid() {
		panic("plow: invalid http method")
	}
	if url == nil {
		panic("plow: passed nil url")
	}
	return &emptyRequest{
		method: method,
		url:    url,
		header: specs.NewHeader(),
	}
}

type emptyRequest struct {
	_ internal.NoCopy

	method specs.HttpMethod
	url    *specs.Url
	header *specs.Header
}

func (req *emptyRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *emptyRequest) Url() *specs.Url {
	return req.url
}

func (req *emptyRequest) Header() *specs.Header {
	return req.header
}

// TextRequest is implementation for the [ClientRequest]
// with string as request body to be sent by the [Client] or [Transport].
//
// Content type applies as "Content-Type" header value
//
// if method unspecified then [specs.HttpMethodPost] will be set
// if content type unspecified then [specs.ContentTypePlain] will be set
func TextRequest(method specs.HttpMethod, url *specs.Url, contentType string, text string) ClientRequest {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return BufferRequest(method, url, contentType, []byte(text))
}

// BufferRequest is implementation for the [ClientRequest]
// with []byte as request body to be sent by the [Client] or [Transport].
//
// Content type applies as "Content-Type" header value
//
// if method unspecified then [specs.HttpMethodPost] will be set
// if content type unspecified then [specs.ContentTypeRaw] will be set
func BufferRequest(method specs.HttpMethod, url *specs.Url, contentType string, buffer []byte) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsPostable() {
		panic(fmt.Sprintf("plow: http method '%s' is not postable", method))
	}
	if buffer == nil {
		panic("plow: passed nil buffer")
	}

	req := &bufferRequest{
		ClientRequest: EmptyRequest(method, url),
		buffer:        buffer,
		contentLength: int64(len(buffer)),
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	req.Header().Set("Content-Type", contentType)

	return req
}

type bufferRequest struct {
	ClientRequest
	buffer        []byte
	contentLength int64
}

func (req *bufferRequest) WriteBody(w io.Writer) error {
	_, err := w.Write(req.buffer)
	return err
}

func (req *bufferRequest) ContentLength() int64 {
	return req.contentLength
}

// StreamRequest is implementation for the [ClientRequest] that
// copy response body from [io.Reader] to be sent by the [Client] or [Transport].
//
// Content type applies as "Content-Type" header value
//
// if method unspecified then [specs.HttpMethodPost] will be set
// if content type unspecified then [specs.ContentTypeRaw] will be set
func StreamRequest(method specs.HttpMethod, url *specs.Url, contentType string, stream io.Reader, contentLength int64) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsPostable() {
		panic(fmt.Sprintf("plow: http method '%s' is not postable", method))
	}

	if stream == nil {
		panic("plow: passed nil stream")
	}

	req := &streamRequest{
		ClientRequest: EmptyRequest(method, url),
		stream:        stream,
		contentLength: contentLength,
	}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	req.Header().Set("Content-Type", contentType)

	return req
}

type streamRequest struct {
	ClientRequest
	stream        io.Reader
	contentLength int64
}

func (req *streamRequest) WriteBody(w io.Writer) error {
	_, err := io.Copy(w, req.stream)
	return err
}

func (req *streamRequest) ContentLength() int64 {
	return req.contentLength
}
