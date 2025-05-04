package client

import (
	"github.com/oesand/giglet/specs"
	"io"
)

type HttpClientResponse struct {
	status specs.StatusCode
	header *specs.Header
	body   io.ReadCloser
}

func (resp *HttpClientResponse) StatusCode() specs.StatusCode {
	return resp.status
}

func (resp *HttpClientResponse) Header() *specs.Header {
	return resp.header
}

func (resp *HttpClientResponse) SetBody(val io.ReadCloser) {
	resp.body = val
}

func (resp *HttpClientResponse) Body() io.ReadCloser {
	return resp.body
}
