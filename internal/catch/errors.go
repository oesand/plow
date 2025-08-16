package catch

import (
	"context"
	"errors"
	"github.com/oesand/giglet/specs"
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

func CatchContextCancel(ctx context.Context) error {
	err := CatchCommonErr(ctx.Err())
	if err == nil {
		return nil
	}
	return TryWrapOpErr("cause", err)
}

func TryWrapOpErr(op specs.GigletOp, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, specs.ErrCancelled) ||
		errors.Is(err, specs.ErrTimeout) ||
		errors.Is(err, specs.ErrClosed) {
		return err
	}
	if _, ok := err.(*specs.GigletError); ok {
		return err
	}
	return &specs.GigletError{
		Op:  op,
		Err: err,
	}
}
