package specs

var (
	ErrCancelled     = NewOpError("context", "cancelled")
	ErrInvalidFormat = NewOpError("parsing", "invalid format")
	ErrTooLarge      = NewOpError("read", "too large content")
)
