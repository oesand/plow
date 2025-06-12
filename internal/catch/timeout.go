package catch

import (
	"context"
	"errors"
	"github.com/oesand/giglet/specs"
	"time"
)

type timeoutRes[T any] struct {
	res T
	err error
}

func CallWithTimeoutContext[TR any](ctx context.Context, timeout time.Duration, fn func(context.Context) (TR, error)) (TR, error) {
	if timeout <= 0 {
		return fn(ctx)
	}

	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	resc := make(chan timeoutRes[TR], 1)

	go func(ctx context.Context, resc chan timeoutRes[TR]) {
		res, err := fn(ctx)
		resc <- timeoutRes[TR]{res, err}
	}(ctx, resc)

	select {
	case <-ctx.Done():
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			err = specs.ErrTimeout
		}
		if errors.Is(err, context.Canceled) {
			err = specs.ErrCancelled
		}
		return *new(TR), err
	case res := <-resc:
		return res.res, res.err
	}
}
