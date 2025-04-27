package giglet

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"net"
	"time"
)

type Handler func(request Request) Response
type HijackHandler = server.HijackHandler
type NextProtoHandler func(conn *tls.Conn)
type ConnHandler func(addr net.Conn, context context.Context) context.Context
type EventHandler func()

const DefaultServerName = "giglet"

var (
	ErrorCancelled = &specs.GigletError{Err: errors.New("cancelled")}

	zeroDialer         net.Dialer
	zeroTime           time.Time
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
