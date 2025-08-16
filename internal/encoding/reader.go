package encoding

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/oesand/plow/specs"
	"io"
)

func NewReader(contentEncoding string, reader io.Reader) (io.ReadCloser, error) {
	switch contentEncoding {
	case specs.ContentEncodingGzip:
		return gzip.NewReader(reader)
	case specs.ContentEncodingDeflate:
		return zlib.NewReader(reader)
	case specs.ContentEncodingBrotli:
		return io.NopCloser(brotli.NewReader(reader)), nil
	}
	return nil, fmt.Errorf("unknown content encoding %s", contentEncoding)
}
