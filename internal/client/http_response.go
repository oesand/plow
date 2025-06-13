package client

import (
	"github.com/oesand/giglet/specs"
	"io"
)

type HttpClientResponse struct {
	status specs.StatusCode
	header *specs.Header

	Reader   io.ReadCloser
	Hijacked bool
}

func (resp *HttpClientResponse) StatusCode() specs.StatusCode {
	return resp.status
}

func (resp *HttpClientResponse) Header() *specs.Header {
	return resp.header
}

func (resp *HttpClientResponse) Body() io.ReadCloser {
	return resp.Reader
}
