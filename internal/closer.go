package internal

import "io"

func ReadCloser(reader io.Reader, closer io.Closer) io.ReadCloser {
	return &readCloser{
		Reader: reader,
		Closer: closer,
	}
}

type readCloser struct {
	io.Reader
	io.Closer
}

type CloserFunc func() error

func (f CloserFunc) Close() error { return f() }
