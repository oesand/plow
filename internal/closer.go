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

func NopWCloser(w io.Writer) io.WriteCloser {
	return nopWCloser{w}
}

type nopWCloser struct {
	io.Writer
}

func (nopWCloser) Close() error { return nil }

type CloserFunc func() error

func (f CloserFunc) Close() error { return f() }
