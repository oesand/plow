package specs

import (
	"errors"
)

var (
	ErrClosed        = errors.New("closed")
	ErrTimeout       = NewOpError("conn", "timeout")
	ErrCancelled     = NewOpError("context", "cancelled")
	ErrInvalidFormat = NewOpError("parsing", "invalid format")
	ErrTooLarge      = NewOpError("read", "too large content")
)
