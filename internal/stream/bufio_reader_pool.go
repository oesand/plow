package stream

import (
	"bufio"
	"io"
	"sync"
)

var DefaultBufioReaderPool = BufioReaderPool{MaxSize: 1024}

type BufioReaderPool struct {
	pool    sync.Pool
	MaxSize int
}

func (rdp *BufioReaderPool) Get(reader io.Reader) *bufio.Reader {
	if item := rdp.pool.Get(); item != nil {
		rd := item.(*bufio.Reader)
		rd.Reset(reader)
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
