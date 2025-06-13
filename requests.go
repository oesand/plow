package giglet

import (
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
)

func NewRequest(method specs.HttpMethod, url *specs.Url) ClientRequest {
	return newRequest(method, url)
}

func newRequest(method specs.HttpMethod, url *specs.Url) *clientRequest {
	return &clientRequest{
		method: method,
		url:    *url,
		header: specs.NewHeader(),
	}
}

type clientRequest struct {
	_ utils.NoCopy

	method specs.HttpMethod
	url    specs.Url
	header *specs.Header
}

func (req *clientRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *clientRequest) Url() specs.Url {
	return req.url
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
		clientRequest: *newRequest(method, url),
		buffer:        buffer,
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
		clientRequest: *newRequest(method, url),
		stream:        stream,
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

func NewHijackRequest(method specs.HttpMethod, url *specs.Url) HijackRequest {
	return &hijackRequest{
		clientRequest: *newRequest(method, url),
	}
}

type hijackRequest struct {
	clientRequest
	conn net.Conn
}

func (req *hijackRequest) Hijack(conn net.Conn) {
	req.conn = conn
}

func (req *hijackRequest) Conn() net.Conn {
	return req.conn
}
