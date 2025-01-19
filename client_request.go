package giglet

import (
	"giglet/internal"
	"giglet/specs"
	"io"
)

func NewRequest(method specs.HttpMethod, url *specs.Url) *HttpClientRequest {
	return NewPostRequest(method, url, nil)
}

func NewPostRequest(method specs.HttpMethod, url *specs.Url, buffer []byte) *HttpClientRequest {
	if method == "" {
		if buffer == nil {
			method = specs.HttpMethodGet
		} else {
			method = specs.HttpMethodPost
		}
	} else if !method.IsValid() {
		panic("giglet/request: invalid method")
	}
	if url == nil {
		panic("giglet/request: invalid url")
	}

	return &HttpClientRequest{
		method: method,
		url:    url,
		header: &specs.Header{},
		buffer: buffer,
	}
}

func NewPostStreamRequest(method specs.HttpMethod, url *specs.Url, stream io.Reader) *HttpClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsValid() {
		panic("giglet/request: invalid method")
	}
	if url == nil {
		panic("giglet/request: invalid url")
	}
	if stream == nil {
		panic("giglet/request: invalid stream")
	}

	return &HttpClientRequest{
		method: method,
		url:    url,
		header: &specs.Header{},
		stream: stream,
	}
}

type HttpClientRequest struct {
	_ internal.NoCopy

	method specs.HttpMethod
	url    *specs.Url
	header *specs.Header

	buffer []byte
	stream io.Reader
}

func (req *HttpClientRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *HttpClientRequest) Url() specs.Url {
	return *req.url
}

func (req *HttpClientRequest) Header() *specs.Header {
	return req.header
}

func (req *HttpClientRequest) WriteBody(writer io.Writer) (err error) {
	if req.buffer == nil {
		_, err = writer.Write(req.buffer)
	} else if req.stream != nil {
		_, err = io.Copy(writer, req.stream)
	}

	return
}
