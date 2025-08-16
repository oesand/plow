package specs

import (
	"errors"
	"fmt"
)

// NewOpError creates a new OpError with the specified operation and formatted error message.
func NewOpError(op OpName, format string, a ...any) error {
	return &OpError{
		Op:  op,
		Err: fmt.Errorf(format, a...),
	}
}

// OpName represents an operation in the plow.
type OpName string

// OpError represents an error that occurred during a plow operation.
type OpError struct {
	Op  OpName
	Err error
}

// String formats the OpError as a string, including the operation if it exists.
func (e *OpError) String() string {
	if e.Op != "" {
		return fmt.Sprintf("plow/%s: %s", e.Op, e.Err)
	}
	return fmt.Sprintf("plow: %s", e.Err)
}

// Error implements the error interface for OpError.
func (e *OpError) Error() string {
	return e.String()
}

// Match checks if the OpError matches another error based on operation and underlying error.
func (e *OpError) Match(err error) bool {
	var oerr *OpError
	if errors.As(err, &oerr) {
		return e != nil && oerr != nil && e.Op == oerr.Op && errors.Is(e.Err, oerr.Err)
	}
	return errors.Is(e, err)
}
