package specs

import (
	"errors"
	"fmt"
)

// NewOpError creates a new GigletError with the specified operation and formatted error message.
func NewOpError(op GigletOp, format string, a ...any) error {
	return &GigletError{
		Op:  op,
		Err: fmt.Errorf(format, a...),
	}
}

// GigletOp represents an operation in the giglet.
type GigletOp string

// GigletError represents an error that occurred during a giglet operation.
type GigletError struct {
	Op  GigletOp
	Err error
}

// String formats the GigletError as a string, including the operation if it exists.
func (e *GigletError) String() string {
	if e.Op != "" {
		return fmt.Sprintf("giglet/%s: %s", e.Op, e.Err)
	}
	return fmt.Sprintf("giglet: %s", e.Err)
}

// Error implements the error interface for GigletError.
func (e *GigletError) Error() string {
	return e.String()
}

// Match checks if the GigletError matches another error based on operation and underlying error.
func (e *GigletError) Match(err error) bool {
	var oerr *GigletError
	if errors.As(err, &oerr) {
		return e != nil && oerr != nil && e.Op == oerr.Op && errors.Is(e.Err, oerr.Err)
	}
	return errors.Is(e, err)
}
