package responses

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
)

type HeaderResponse struct {
	_ internal.NoCopy

	statusCode specs.StatusCode
	header     *specs.Header
}

func (resp *HeaderResponse) StatusCode() specs.StatusCode {
	if resp.statusCode == specs.StatusCodeUndefined {
		resp.statusCode = specs.StatusCodeOK
	}
	return resp.statusCode
}

func (resp *HeaderResponse) SetStatusCode(code specs.StatusCode) {
	resp.statusCode = code
}

func (resp *HeaderResponse) Header() *specs.Header {
	if resp.header == nil {
		resp.header = &specs.Header{}
	}
	return resp.header
}
