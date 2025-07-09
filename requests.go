package giglet

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
)

func NewRequest(method specs.HttpMethod, url *specs.Url) ClientRequest {
	return newRequest(method, url)
}

func newRequest(method specs.HttpMethod, url *specs.Url) *clientRequest {
	if !method.IsValid() {
		panic("giglet/request: invalid method")
	}
	if url == nil {
		panic("giglet/request: passed nil url")
	}
	return &clientRequest{
		method: method,
		url:    *url,
		header: specs.NewHeader(),
	}
}

type clientRequest struct {
	_ internal.NoCopy

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

func NewTextRequest(method specs.HttpMethod, url *specs.Url, text string, contentType string) ClientRequest {
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	return NewBufferRequest(method, url, []byte(text), contentType)
}

func NewBufferRequest(method specs.HttpMethod, url *specs.Url, buffer []byte, contentType string) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	}

	req := &bufferRequest{
		clientRequest: *newRequest(method, url),
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
	clientRequest
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

func NewStreamRequest(method specs.HttpMethod, url *specs.Url, stream io.Reader, contentType string, contentLength int64) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	}
	if contentLength < 0 {
		panic("giglet/request: invalid content length")
	}

	req := &streamRequest{
		clientRequest: *newRequest(method, url),
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
	clientRequest
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
