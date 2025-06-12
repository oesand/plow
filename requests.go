package giglet

import (
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
	"io"
)

func NewRequest(method specs.HttpMethod, url *specs.Url) ClientRequest {
	return &clientRequest{
		method: method,
		url:    url,
		header: &specs.Header{},
	}
}

type clientRequest struct {
	_ utils.NoCopy

	method specs.HttpMethod
	url    *specs.Url
	header *specs.Header
}

func (req *clientRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *clientRequest) Url() specs.Url {
	return *req.url
}

func (req *clientRequest) Header() *specs.Header {
	return req.header
}

func NewBufferRequest(method specs.HttpMethod, url *specs.Url, buffer []byte) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsValid() {
		panic("giglet/request: invalid method")
	}
	if url == nil {
		panic("giglet/request: invalid url")
	}

	return &bufferRequest{
		clientRequest: clientRequest{
			method: method,
			url:    url,
			header: specs.NewHeader(),
		},
		buffer: buffer,
	}
}

type bufferRequest struct {
	clientRequest
	buffer []byte
}

func (req *bufferRequest) WriteBody(w io.Writer) error {
	_, err := w.Write(req.buffer)
	return err
}

func NewStreamRequest(method specs.HttpMethod, url *specs.Url, stream io.Reader) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsValid() {
		panic("giglet/request: invalid method")
	}
	if url == nil {
		panic("giglet/request: invalid url")
	}

	return &streamRequest{
		clientRequest: clientRequest{
			method: method,
			url:    url,
			header: specs.NewHeader(),
		},
		stream: stream,
	}
}

type streamRequest struct {
	clientRequest
	stream io.Reader
}

func (req *streamRequest) WriteBody(w io.Writer) error {
	_, err := io.Copy(w, req.stream)
	return err
}
