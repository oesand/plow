package internal

import (
	"io"
	"sync/atomic"
)

type SeqCloser struct {
	closers []io.Closer
	closed  atomic.Bool
}

func (rc *SeqCloser) Add(closer io.Closer) {
	if closer == nil {
		panic("closer cannot be nil")
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
	for c := range ReverseIter(rc.closers) {
		cerr := c.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}
