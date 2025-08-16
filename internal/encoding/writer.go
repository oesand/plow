package encoding

import (
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/oesand/giglet/specs"
	"io"
)

func NewWriter(contentEncoding string, writer io.Writer) (io.WriteCloser, error) {
	switch contentEncoding {
	case specs.ContentEncodingGzip:
		return gzip.NewWriter(writer), nil
	case specs.ContentEncodingDeflate:
		return zlib.NewWriter(writer), nil
	case specs.ContentEncodingBrotli:
		return brotli.NewWriter(writer), nil
	}
	return nil, fmt.Errorf("unknown content encoding %s", contentEncoding)
}
