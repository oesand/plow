package internal

import (
	"io"
	"sync/atomic"
)

func ReadClose(reading Reading, closing Closing) io.ReadCloser {
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
	reading Reading
	closing Closing
}

func (comb *readClose) Read(p []byte) (int, error) {
	if comb.closed.Load() {
		return -1, ErrorAlreadyClosed
	}
	return comb.reading(p)
}

func (comb *readClose) Close() error {
	if comb.closing != nil {
		if comb.closed.Load() {
			return ErrorAlreadyClosed
		}
		defer comb.closed.Store(true)
		return comb.closing()
	}
	return nil
}
