package encoding

import (
	"bufio"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/oesand/plow/specs"
	"io"
)

func NewReader(isChunked bool, contentEncoding string, bufio *bufio.Reader) (io.ReadCloser, error) {
	var reader io.Reader
	if isChunked {
		reader = NewChunkedReader(bufio)
	} else {
		reader = bufio
	}

	switch contentEncoding {
	case "":
		return io.NopCloser(reader), nil
	case specs.ContentEncodingGzip:
		return gzip.NewReader(reader)
	case specs.ContentEncodingDeflate:
		return zlib.NewReader(reader)
	case specs.ContentEncodingBrotli:
		return io.NopCloser(brotli.NewReader(reader)), nil
	}
	return nil, fmt.Errorf("unknown content encoding %s", contentEncoding)
}
