package internal

import (
	"errors"
)

var (
	ErrorTooLarge      = errors.New("giglet: too large data")
	ErrorAlreadyClosed = errors.New("giglet: already closed")
)

type Reading func(p []byte) (int, error)
type Closing func() error
