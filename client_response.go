package giglet

import (
	"github.com/oesand/giglet/specs"
	"io"
)

type HttpClientResponse struct {
	status specs.StatusCode
	header *specs.ReadOnlyHeader
	body   io.ReadCloser
}

func (resp *HttpClientResponse) StatusCode() specs.StatusCode {
	return resp.status
}

func (resp *HttpClientResponse) Header() *specs.ReadOnlyHeader {
	return resp.header
}

func (resp *HttpClientResponse) Body() io.ReadCloser {
	return resp.body
}
