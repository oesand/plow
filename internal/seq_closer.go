package internal

import (
	"io"
	"slices"
	"sync/atomic"
)

type SeqCloser struct {
	closers []io.Closer
	closed  atomic.Bool
}

func (rc *SeqCloser) Add(closer io.Closer) {
	if closer == nil {
		panic("plow: closer cannot be nil")
	}
	rc.closers = append(rc.closers, closer)
}

func (rc *SeqCloser) Close() error {
	if rc.closed.Load() {
		return nil
	}
	rc.closed.Store(true)

	if len(rc.closers) == 0 {
		return nil
	}

	var err error
	for _, c := range slices.Backward(rc.closers) {
		cerr := c.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}
