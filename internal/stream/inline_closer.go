package stream

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"sync/atomic"
)

func Closer(closing internal.Closing) io.Closer {
	return &closer{
		closing: closing,
	}
}

type closer struct {
	closed  atomic.Bool
	closing internal.Closing
}

func (comb *closer) Close() error {
	if comb.closing != nil {
		if comb.closed.Load() {
			return specs.ErrClosed
		}
		defer comb.closed.Store(true)
		return comb.closing()
	}
	return nil
}
