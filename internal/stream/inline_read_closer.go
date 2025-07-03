package stream

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"sync/atomic"
)

func ReadClose(reading internal.Reading, closing internal.Closing) io.ReadCloser {
	if reading == nil {
		panic("giglet/internal: reader cannot be empty")
	}

	return &readClose{
		reading: reading,
		closing: closing,
	}
}

type readClose struct {
	closed  atomic.Bool
	reading internal.Reading
	closing internal.Closing
}

func (comb *readClose) Read(p []byte) (int, error) {
	if comb.closed.Load() {
		return -1, specs.ErrClosed
	}
	return comb.reading(p)
}

func (comb *readClose) Close() error {
	if comb.closing != nil {
		if comb.closed.Load() {
			return specs.ErrClosed
		}
		defer comb.closed.Store(true)
		return comb.closing()
	}
	return nil
}
