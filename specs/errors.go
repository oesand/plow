package specs

import (
	"errors"
)

var (
	ErrClosed                  = errors.New("closed")
	ErrProtocol                = errors.New("protocol implementation error")
	ErrTimeout                 = errors.New("i/o timeout")
	ErrCancelled               = NewOpError("context", "cancelled")
	ErrTooLarge                = NewOpError("read", "too large content")
	ErrUnknownTransferEncoding = NewOpError("http", "unknown transfer encoding")
	ErrUnknownContentEncoding  = NewOpError("http", "unknown content encoding")
)
