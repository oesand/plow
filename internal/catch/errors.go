package catch

import (
	"context"
	"errors"
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/specs"
	"io"
	"net"
)

func IsCommonNetReadError(err error) bool {
	if err == io.EOF {
		return true
	} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return true
	} else if operr, ok := err.(*net.OpError); ok && operr.Op == "read" {
		return true
	}
	return false
}

func CatchCommonErr(err error) error {
	if err == nil {
		return nil
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return specs.ErrTimeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return specs.ErrTimeout
	}
	return err
}

func TryWrapOpErr(op specs.OpName, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, specs.ErrCancelled) ||
		errors.Is(err, specs.ErrTimeout) ||
		errors.Is(err, specs.ErrClosed) {
		return err
	}
	if _, ok := err.(*specs.OpError); ok {
		return err
	}
	if _, ok := err.(*server_ops.ErrorResponse); ok {
		return err
	}
	return &specs.OpError{
		Op:  op,
		Err: err,
	}
}
