package giglet

import (
	"crypto/tls"
	"fmt"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"net"
)

type Handler func(request Request) Response
type HijackHandler = server.HijackHandler
type NextProtoHandler func(conn *tls.Conn)
type EventHandler func()

const DefaultServerName = "giglet"

var (
	zeroDialer         net.Dialer
	httpV1NextProtoTLS = "http/1.1"

	responseErrDowngradeHTTPS = &server.ErrorResponse{
		Code: specs.StatusCodeBadRequest,
		Text: "sent an HTTP request to an HTTPS server.",
	}
	responseErrNotProcessable = &server.ErrorResponse{
		Code: specs.StatusCodeUnprocessableEntity,
		Text: "the request could not be processed.",
	}
)

func validationErr(err string, a ...any) error {
	return &specs.GigletError{
		Op:  "validation",
		Err: fmt.Errorf(err, a...),
	}
}
