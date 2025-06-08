package utils

import (
	"testing"
)

var IsNotTesting = !testing.Testing()

type Reading func(p []byte) (int, error)
type Closing func() error
