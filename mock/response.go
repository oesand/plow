package mock

import (
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal/client_ops"
	"github.com/oesand/giglet/specs"
	"io"
)

// ClientResponse creates a mock client response with the given status code and body.
func ClientResponse(code specs.StatusCode, body io.ReadCloser) giglet.ClientResponse {
	resp := client_ops.NewHttpClientResponse(code, specs.NewHeader())
	resp.Reader = body
	return resp
}
