package stream

import (
	"io"
	"net"
)

func IsCommonNetReadError(err error) bool {
	if err == io.EOF {
		return true
	} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return true
	} else if operr, ok := err.(*net.OpError); ok && operr.Op == "read" {
		return true
	}
	return false
}
