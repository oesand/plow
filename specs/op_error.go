package specs

import (
	"errors"
	"fmt"
)

func NewOpError(op GigletOp, format string, a ...any) *GigletError {
	return &GigletError{
		Op:  op,
		Err: fmt.Errorf(format, a...),
	}
}

type GigletOp string
type GigletError struct {
	Op  GigletOp
	Err error
}

func (e *GigletError) String() string {
	if e.Op != "" {
		return fmt.Sprintf("giglet/%s: %s", e.Op, e.Err)
	}
	return fmt.Sprintf("giglet: %s", e.Err)
}

func (e *GigletError) Error() string {
	return e.String()
}

func (e *GigletError) Match(err error) bool {
	var oerr *GigletError
	if errors.As(err, &oerr) {
		return e != nil && oerr != nil && e.Op == oerr.Op && errors.Is(e.Err, oerr.Err)
	}
	return errors.Is(e, err)
}
