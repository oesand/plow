package parsing

import (
	"errors"
	"github.com/oesand/plow/specs"
	"strconv"
)

var ErrParsing = errors.New("cannot parse value")

func ParseContentLength(header *specs.Header) (isChunked bool, size int64, err error) {
	if te, has := header.TryGet("Transfer-Encoding"); has {
		switch te {
		case "chunked":
			isChunked = true
		default:
			err = specs.ErrUnknownTransferEncoding
			return
		}
	} else if cl := header.Get("Content-Length"); cl != "" {
		size, err = strconv.ParseInt(cl, 10, 64)
		if err != nil {
			err = ErrParsing
		}
	}
	return
}
