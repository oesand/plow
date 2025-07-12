package encoding

import (
	"fmt"
	"io"
)

func NewChunkedWriter(writer io.Writer) io.WriteCloser {
	return &chunkedWriter{
		writer: writer,
	}
}

type chunkedWriter struct {
	writer io.Writer
}

func (cw *chunkedWriter) Write(data []byte) (n int, err error) {
	// Don't send 0-length data. It looks like EOF for chunked encoding.
	if len(data) == 0 {
		return 0, nil
	}

	if _, err = fmt.Fprintf(cw.writer, "%x\r\n", len(data)); err != nil {
		return 0, err
	}
	if n, err = cw.writer.Write(data); err != nil {
		return
	}
	if n != len(data) {
		err = io.ErrShortWrite
		return
	}
	if _, err = io.WriteString(cw.writer, "\r\n"); err != nil {
		return
	}
	return
}

func (cw *chunkedWriter) Close() error {
	_, err := io.WriteString(cw.writer, "0\r\n\r\n")
	return err
}
