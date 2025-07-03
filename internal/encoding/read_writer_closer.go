package encoding

import (
	"github.com/oesand/giglet/internal"
	"io"
)

func newReaderCloser(reader io.Reader, chainedClosers []io.Closer) io.ReadCloser {
	return &readWriteCloser{
		Reader:         reader,
		chainedClosers: chainedClosers,
	}
}

func newWriterCloser(writer io.Writer, chainedClosers []io.Closer) io.WriteCloser {
	return &readWriteCloser{
		Writer:         writer,
		chainedClosers: chainedClosers,
	}
}

type readWriteCloser struct {
	io.Reader
	io.Writer
	chainedClosers []io.Closer
	closed         bool
}

func (rc *readWriteCloser) Close() error {
	if rc.closed {
		return nil
	}
	rc.closed = true

	var err error
	for c := range internal.ReverseIter(rc.chainedClosers) {
		cerr := c.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}
