package giglet

import (
	"context"
	"errors"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"mime/multipart"
	"net"
	"net/url"
)

type Request interface {
	Context() context.Context
	PutContext(context context.Context)

	ProtoAtLeast(major, minor uint16) bool
	ProtoNoHigher(major, minor uint16) bool
	RemoteAddr() net.Addr
	Hijack(handler HijackHandler)

	Method() specs.HttpMethod
	Url() *specs.Url
	Header() *specs.ReadOnlyHeader

	Read([]byte) (int, error)
	PostBody() ([]byte, error)
	PostForm() (specs.Form, error)
	MultipartForm() (*multipart.Form, error)
}

type httpRequest struct {
	_ internal.NoCopy

	server   *Server
	conn     net.Conn
	hijacker HijackHandler
	context  context.Context

	protoMajor, protoMinor uint16
	method                 specs.HttpMethod
	url                    *specs.Url
	header                 *specs.ReadOnlyHeader

	body            io.Reader
	bodyReaded      bool
	cachedBody      []byte
	cachedMultipart *multipart.Form
	cachedForm      specs.Form
}

func (req *httpRequest) ProtoAtLeast(major, minor uint16) bool {
	return req.protoMajor > major ||
		req.protoMajor == major && req.protoMinor >= minor
}

func (req *httpRequest) ProtoNoHigher(major, minor uint16) bool {
	return req.protoMajor < major ||
		req.protoMajor == major && req.protoMinor <= minor
}

func (req *httpRequest) Read(buf []byte) (n int, err error) {
	if req.body == nil || req.bodyReaded {
		return 0, io.EOF
	}
	if req.server != nil {
		req.server.applyReadTimeout(req.conn)
		defer req.conn.SetDeadline(zeroTime)
	}

	n, err = req.body.Read(buf)
	if err == io.EOF {
		req.bodyReaded = true
	}
	return n, err
}

func (req *httpRequest) PostBody() (buf []byte, err error) {
	if req.body == nil || (req.bodyReaded && req.cachedBody == nil) {
		return nil, io.EOF
	} else if req.cachedBody != nil {
		return req.cachedBody, nil
	}

	buf, err = io.ReadAll(req)
	if err == nil {
		req.cachedBody = buf
	}

	return
}

func (req *httpRequest) RemoteAddr() net.Addr {
	return req.conn.RemoteAddr()
}

func (req *httpRequest) Hijack(handler HijackHandler) {
	req.hijacker = handler
}

func (req *httpRequest) Method() specs.HttpMethod {
	return req.method
}

func (req *httpRequest) Url() *specs.Url {
	return req.url
}

func (req *httpRequest) Header() *specs.ReadOnlyHeader {
	return req.header
}

func (req *httpRequest) Context() context.Context {
	return req.context
}

func (req *httpRequest) PutContext(context context.Context) {
	req.context = context
}

func (req *httpRequest) PostForm() (specs.Form, error) {
	if req.body == nil {
		return nil, io.EOF
	} else if req.cachedForm != nil {
		return req.cachedForm, nil
	} else if req.Header().ContentType() != specs.ContentTypeMultipart {
		_, err := req.MultipartForm()
		if err != nil {
			return nil, err
		}
		return req.cachedForm, nil
	} else if req.Header().ContentType() != specs.ContentTypeForm {
		return nil, errors.New("giglet: this Content-Type is not a urlencoded-form")
	} else if req.bodyReaded {
		return nil, nil
	}
	req.bodyReaded = true

	buf, err := io.ReadAll(req)
	if err != nil {
		return nil, err
	}
	values, err := url.ParseQuery(string(buf))
	req.cachedForm = specs.Form(values)
	if err != nil {
		return nil, err
	}
	return req.cachedForm, nil
}

func (req *httpRequest) MultipartForm() (*multipart.Form, error) {
	if req.body == nil {
		return nil, io.EOF
	} else if req.Header().ContentType() != specs.ContentTypeMultipart {
		return nil, errors.New("giglet: this Content-Type is not a multipart-form")
	} else if req.cachedMultipart != nil {
		return req.cachedMultipart, nil
	} else if req.bodyReaded {
		return nil, nil
	}
	req.bodyReaded = true

	boundary := req.header.GetMediaParams("boundary")
	if len(boundary) == 0 {
		return nil, errors.New("giglet: this request Content-Type does not contains boundary")
	}

	reader := multipart.NewReader(req, boundary)
	form, err := reader.ReadForm(0)
	if err != nil {
		return nil, err
	}
	req.cachedMultipart = form
	req.cachedForm = req.cachedMultipart.Value
	return req.cachedMultipart, nil
}
