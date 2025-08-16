package mock

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/internal/client_ops"
	"github.com/oesand/plow/specs"
	"io"
)

// ClientResponse creates a mock client response with the given status code and body.
func ClientResponse(code specs.StatusCode, body io.ReadCloser) plow.ClientResponse {
	resp := client_ops.NewHttpClientResponse(code, specs.NewHeader())
	resp.Reader = body
	return resp
}
