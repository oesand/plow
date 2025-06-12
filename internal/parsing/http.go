package parsing

import (
	"bytes"
	"github.com/oesand/giglet/internal/utils"
	"strconv"
)

var (
	httpVersionPrefix = []byte("HTTP/")
	httpV10           = []byte("HTTP/1.0")
	httpV11           = []byte("HTTP/1.1")
	httpV2            = []byte("HTTP/2.0")
)

func parseHTTPVersion(value []byte) (major, minor uint16, ok bool) {
	if bytes.EqualFold(value, httpV10) {
		return 1, 0, true
	} else if bytes.EqualFold(value, httpV11) {
		return 1, 1, true
	} else if bytes.EqualFold(value, httpV2) {
		return 2, 0, true
	} else if !bytes.HasPrefix(value, httpVersionPrefix) ||
		len(value) != 8 || value[6] != '.' {
		return 0, 0, false
	}

	maj, err := strconv.ParseUint(utils.BufferToString(value[5:6]), 10, 16)
	if err != nil {
		return 0, 0, false
	}
	min, err := strconv.ParseUint(utils.BufferToString(value[7:8]), 10, 16)
	if err != nil {
		return 0, 0, false
	}
	return uint16(maj), uint16(min), true
}
