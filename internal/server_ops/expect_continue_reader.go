package server_ops

import (
	"io"
	"sync/atomic"
)

var ResponseContinueBuf = []byte("HTTP/1.1 100 Continue\r\n\r\n")

func ExpectContinueReader(bodyReader io.Reader, writer io.Writer) io.Reader {
	return &expectContinueReader{
		bodyReader: bodyReader,
		writer:     writer,
	}
}

type expectContinueReader struct {
	bodyReader      io.Reader
	writer          io.Writer
	writtenContinue atomic.Bool
}

func (r *expectContinueReader) Read(b []byte) (int, error) {
	if !r.writtenContinue.Load() {
		r.writtenContinue.Store(true)

		_, err := r.writer.Write(ResponseContinueBuf)
		if err != nil {
			return 0, err
		}
	}
	return r.bodyReader.Read(b)
}
