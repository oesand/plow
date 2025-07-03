package specs

import (
	"errors"
)

var (
	ErrClosed                  = errors.New("closed")
	ErrTimeout                 = errors.New("timeout")
	ErrCancelled               = NewOpError("context", "cancelled")
	ErrInvalidFormat           = NewOpError("parsing", "invalid format")
	ErrTooLarge                = NewOpError("read", "too large content")
	ErrUnknownTransferEncoding = NewOpError("http", "unknown transfer encoding")
	ErrUnknownContentEncoding  = NewOpError("http", "unknown content encoding")
)
