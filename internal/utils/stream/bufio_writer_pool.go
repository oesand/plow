package stream

import (
	"bufio"
	"io"
	"sync"
)

var DefaultBufioWriterPool BufioWriterPool

type BufioWriterPool struct {
	pool    sync.Pool
	MaxSize int
}

func (rdp *BufioWriterPool) Get(writer io.Writer) *bufio.Writer {
	if item := rdp.pool.Get(); item != nil {
		wr := item.(*bufio.Writer)
		wr.Reset(writer)
	}
	if rdp.MaxSize > 0 {
		return bufio.NewWriterSize(writer, rdp.MaxSize)
	}
	return bufio.NewWriter(writer)
}

func (rdp *BufioWriterPool) Put(writer *bufio.Writer) {
	writer.Reset(nil)
	rdp.pool.Put(writer)
}
