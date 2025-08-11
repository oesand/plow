package ws

import "io"

type frameData struct {
	io.Reader

	Fin    bool
	Code   wsFrameType
	Length int
}

func unmaskReader(reader io.Reader, maxSize int64, maskingKey []byte) io.Reader {
	return &unmaskingReader{
		LimitedReader: io.LimitedReader{R: reader, N: maxSize},
		maskingKey:    maskingKey,
	}
}

type unmaskingReader struct {
	io.LimitedReader

	maskingKey []byte
	pos        int64
}

func (reader *unmaskingReader) Read(p []byte) (int, error) {
	n, err := reader.LimitedReader.Read(p)
	if err == nil && reader.maskingKey != nil {
		for i := 0; i < n; i++ {
			p[i] ^= reader.maskingKey[reader.pos%4]
			reader.pos++
		}
	}
	return n, err
}
