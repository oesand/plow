package catch

import (
	"context"
	"time"
)

type timeoutResErr[T any] struct {
	res T
	err error
}

func CallWithTimeoutContext[TR any](ctx context.Context, timeout time.Duration, fn func(context.Context) (TR, error)) (TR, error) {
	if timeout <= 0 {
		return fn(ctx)
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	resc := make(chan timeoutResErr[TR], 1)

	go func(ctx context.Context, ch chan timeoutResErr[TR]) {
		res, err := fn(ctx)
		ch <- timeoutResErr[TR]{res, err}
	}(ctx, resc)

	select {
	case <-ctx.Done():
		err := ctx.Err()
		return *new(TR), CatchCommonErr(err)
	case res := <-resc:
		return res.res, CatchCommonErr(res.err)
	}
}

func CallWithTimeoutContextErr(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if timeout <= 0 {
		return fn(ctx)
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

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
