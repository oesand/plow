package catch

import (
	"context"
	"time"
)

type ResultErrPair[T any] struct {
	Res T
	Err error
}

func CallWithTimeoutContext[TR any](ctx context.Context, timeout time.Duration, fn func(context.Context) (TR, error)) (TR, error) {
	if timeout > 0 {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
		defer cancelTimeout()
	}

	resc := make(chan ResultErrPair[TR], 1)

	go func(ctx context.Context, ch chan ResultErrPair[TR]) {
		res, err := fn(ctx)
		ch <- ResultErrPair[TR]{res, err}
	}(ctx, resc)

	select {
	case <-ctx.Done():
		err := ctx.Err()
		return *new(TR), CatchCommonErr(err)
	case res := <-resc:
		return res.Res, CatchCommonErr(res.Err)
	}
}

func CallWithTimeoutContextErr(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if timeout > 0 {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, timeout)
		defer cancelTimeout()
	}

	errch := make(chan error, 1)

	go func(ctx context.Context, ch chan error) {
		err := fn(ctx)
		ch <- err
	}(ctx, errch)

	select {
	case <-ctx.Done():
		err := ctx.Err()
		return CatchCommonErr(err)
	case err := <-errch:
		return CatchCommonErr(err)
	}
}
