package encoding

import (
	"compress/gzip"
	"github.com/oesand/giglet/specs"
	"io"
	"net/http/httputil"
)

type ReadWriterBuilder struct {
	Encoding string
	Chunked  bool
}

func (builder ReadWriterBuilder) NewReader(reader io.Reader, afterClosers ...io.Closer) (io.ReadCloser, error) {
	var chainedClosers = afterClosers

	if builder.Chunked {
		reader = httputil.NewChunkedReader(reader)
	}

	switch builder.Encoding {
	case specs.GzipContentEncoding:
		gzr, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		reader = gzr
		chainedClosers = append(chainedClosers, gzr)
	}

	return newReaderCloser(reader, chainedClosers), nil
}

func (builder ReadWriterBuilder) NewWriter(writer io.Writer, afterClosers ...io.Closer) (io.WriteCloser, error) {
	var chainedClosers = afterClosers

	if builder.Chunked {
		chw := httputil.NewChunkedWriter(writer)
		writer = chw
		chainedClosers = append(chainedClosers, chw)
	}

	switch builder.Encoding {
	case specs.GzipContentEncoding:
		gzw := gzip.NewWriter(writer)
		writer = gzw
		chainedClosers = append(chainedClosers, gzw)
	}

	return newWriterCloser(writer, chainedClosers), nil
}
