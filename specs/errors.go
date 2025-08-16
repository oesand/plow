package specs

import (
	"context"
	"errors"
)

var (
	ErrClosed                  = errors.New("closed")
	ErrTimeout                 = errors.New("timeout")
	ErrCancelled               = context.Canceled
	ErrProtocol                = NewOpError("protocol", "implementation error")
	ErrTooLarge                = NewOpError("read", "too large content")
	ErrUnknownTransferEncoding = NewOpError("http", "unknown transfer encoding")
	ErrUnknownContentEncoding  = NewOpError("http", "unknown content encoding")
)
