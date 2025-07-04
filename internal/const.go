package internal

import (
	"io"
	"testing"
)

var IsNotTesting = !testing.Testing()

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
