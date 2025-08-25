package plow

import (
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"time"
)

const (
	// DefaultServerName default value for Server.ServerName parameter
	DefaultServerName = "plow"

	// DefaultMaxRedirectCount default value for Client.MaxRedirectCount parameter
	DefaultMaxRedirectCount int = 10

	// DefaultMaxEncodingSize default value for Server.MaxEncodingSize parameter
	DefaultMaxEncodingSize int64 = 5 << 20 // 5 mb
)

var (
	httpV1NextProtoTLS = "http/1.1"

	defaultDialer = net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	responseContinueBuf = []byte("HTTP/1.1 100 Continue\r\n\r\n")

	responseErrDowngradeHTTPS = &server_ops.ErrorResponse{
		Code: specs.StatusCodeBadRequest,
		Text: "http: sent an HTTP request to an HTTPS server.",
	}
	responseErrNotProcessable = &server_ops.ErrorResponse{
		Code: specs.StatusCodeUnprocessableEntity,
		Text: "http: the request could not be processed.",
	}
	responseErrBodyTooLarge = &server_ops.ErrorResponse{
		Code: specs.StatusCodeRequestEntityTooLarge,
		Text: "http: too large body.",
	}
	responseExpectationFailedError = &server_ops.ErrorResponse{
		Code: specs.StatusCodeExpectationFailed,
	}
	responseInvalidContentLength = &server_ops.ErrorResponse{
		Code: specs.StatusCodeLengthRequired,
		Text: "http: invalid Content-Length header",
	}
	responseUnsupportedTransferEncoding = &server_ops.ErrorResponse{
		Code: specs.StatusCodeNotImplemented,
		Text: "http: unsupported transfer encoding",
	}
	responseInternalServerError = &server_ops.ErrorResponse{
		Code: specs.StatusCodeInternalServerError,
		Text: "http: internal server error",
	}
)

// ShortResponseWriter creates an io.WriterTo implementation that writes a short HTTP error response.
// initialized with the provided status code and text.
//
// This is useful for quickly generating responses in HTTP handlers, ensuring consistent formatting
// and status codes across the application. The returned object implements io.WriterTo, allowing it to be
// written directly to an io.Writer, such as net.Conn.
func ShortResponseWriter(code specs.StatusCode, text string) io.WriterTo {
	return &server_ops.ErrorResponse{
		Code: code,
		Text: text,
	}
}
