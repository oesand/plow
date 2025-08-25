package encoding

import (
	"bufio"
	"bytes"
	"github.com/oesand/plow/specs"
	"io"
	"net/http/httputil"
	"sync"
)

var singleCRLF = []byte("\r\n")

func NewChunkedReader(buf *bufio.Reader) io.Reader {
	return &chunkedReader{
		chunked: httputil.NewChunkedReader(buf),
		bufio:   buf,
	}
}

type chunkedReader struct {
	chunked io.Reader
	bufio   *bufio.Reader
	mu      sync.Mutex
	sawEOF  bool
}

func (cr *chunkedReader) Read(p []byte) (int, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.sawEOF {
		return 0, io.EOF
	}

	n, err := cr.chunked.Read(p)
	if err == io.EOF {
		cr.sawEOF = true
		buf, err := cr.bufio.Peek(2)
		if bytes.Equal(buf, singleCRLF) {
			cr.bufio.Discard(2)
		} else if len(buf) < 2 {
			return 0, specs.ErrTrailerEOF
		} else if err != nil {
			return 0, err
		}
	}
	return n, err
}
