package internal

import (
	"testing"
)

var IsNotTesting = !testing.Testing()

type FlagKey struct {
	Key string
}
