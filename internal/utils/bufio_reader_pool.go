package utils

import (
	"bufio"
	"io"
	"sync"
)

type BufioReaderPool struct {
	pool    sync.Pool
	MaxSize int
}

func (rdp *BufioReaderPool) Get(reader io.Reader) *bufio.Reader {
	if rd := rdp.pool.Get(); rd != nil {
		reader := rd.(*bufio.Reader)
		reader.Reset(reader)
	}
	if rdp.MaxSize > 0 {
		return bufio.NewReaderSize(reader, rdp.MaxSize)
	}
	return bufio.NewReader(reader)
}

func (rdp *BufioReaderPool) Put(reader *bufio.Reader) {
	reader.Reset(nil)
	rdp.pool.Put(reader)
}
