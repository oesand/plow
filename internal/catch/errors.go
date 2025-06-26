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
	if errors.Is(err, context.Canceled) {
		return specs.ErrCancelled
	}
	return err
}

func CatchContextCancel(ctx context.Context) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return specs.ErrTimeout
	}
	if errors.Is(err, context.Canceled) {
		return specs.ErrCancelled
	}
	if _, ok := err.(*specs.GigletError); !ok {
		return &specs.GigletError{
			Op:  "cause",
			Err: err,
		}
	}
	return err
}
