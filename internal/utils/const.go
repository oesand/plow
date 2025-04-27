package utils

import (
	"errors"
)

var (
	ErrorTooLarge      = errors.New("too large data")
	ErrorAlreadyClosed = errors.New("already closed")
)

type Reading func(p []byte) (int, error)
type Closing func() error
