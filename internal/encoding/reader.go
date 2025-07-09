package encoding

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/oesand/giglet/specs"
	"io"
)

func NewReader(contentEncoding string, reader io.Reader) (io.ReadCloser, error) {
	switch contentEncoding {
	case specs.ContentEncodingGzip:
		return gzip.NewReader(reader)
	case specs.ContentEncodingDeflate:
		return flate.NewReader(reader), nil
	case specs.ContentEncodingBrotli:
		return io.NopCloser(brotli.NewReader(reader)), nil
	}
	return nil, fmt.Errorf("unknown content encoding %s", contentEncoding)
}
