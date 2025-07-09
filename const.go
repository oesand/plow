package giglet

import (
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"net"
)

const (
	// DefaultServerName default value for Server.ServerName parameter
	DefaultServerName = "giglet"

	// DefaultMaxRedirectCount default value for Client.MaxRedirectCount parameter
	DefaultMaxRedirectCount int = 10

	// DefaultMaxEncodingSize default value for Server.MaxEncodingSize parameter
	DefaultMaxEncodingSize int64 = 5 << 20 // 5 mb
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
