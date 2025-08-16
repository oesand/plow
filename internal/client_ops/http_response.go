package client_ops

import (
	"github.com/oesand/plow/specs"
	"io"
)

func NewHttpClientResponse(status specs.StatusCode, header *specs.Header) *HttpClientResponse {
	return &HttpClientResponse{
		status: status,
		header: header,
	}
}

type HttpClientResponse struct {
	status specs.StatusCode
	header *specs.Header

	Reader io.ReadCloser
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
