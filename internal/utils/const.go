package utils

import (
	"errors"
	"testing"
)

var (
	ErrorTooLarge      = errors.New("too large data")
	ErrorAlreadyClosed = errors.New("already closed")

	IsNotTesting = !testing.Testing()
)

type Reading func(p []byte) (int, error)
type Closing func() error
