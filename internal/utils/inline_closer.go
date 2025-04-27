package utils

import (
	"io"
	"sync/atomic"
)

func Closer(closing Closing) io.Closer {
	return &closer{
		closing: closing,
	}
}

type closer struct {
	closed  atomic.Bool
	closing Closing
}

func (comb *closer) Close() error {
	if comb.closing != nil {
		if comb.closed.Load() {
			return ErrorAlreadyClosed
		}
		defer comb.closed.Store(true)
		return comb.closing()
	}
	return nil
}
