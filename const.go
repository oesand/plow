package giglet

import (
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"net"
)

const (
	DefaultServerName           = "giglet"
	DefaultMaxRedirectCount int = 10
)

var (
	zeroDialer         net.Dialer
	httpV1NextProtoTLS = "http/1.1"

	responseErrDowngradeHTTPS = &server.ErrorResponse{
		Code: specs.StatusCodeBadRequest,
		Text: "http: sent an HTTP request to an HTTPS server.",
	}
	responseErrNotProcessable = &server.ErrorResponse{
		Code: specs.StatusCodeUnprocessableEntity,
		Text: "http: the request could not be processed.",
	}
	responseErrBodyTooLarge = &server.ErrorResponse{
		Code: specs.StatusCodeRequestEntityTooLarge,
		Text: "http: too large body",
	}
)
