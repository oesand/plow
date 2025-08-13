package ws

import (
	"bytes"
	"compress/flate"
	"io"
)

func framePayloadReader(rd io.Reader, maxSize int, maskingKey []byte, decompress bool) io.Reader {
	var reader io.Reader = &unmaskingReader{
		LimitedReader: io.LimitedReader{R: rd, N: int64(maxSize)},
		maskingKey:    maskingKey,
	}

	if decompress {
		reader = flate.NewReader(reader)
	}

	return reader
}

func prepareFramePayload(maskingKey []byte, compress bool, payload []byte) ([]byte, error) {
	if compress {
		var buf bytes.Buffer
		fw, err := flate.NewWriter(&buf, flate.BestSpeed)
		if err != nil {
			return nil, err
		}
		if _, err = fw.Write(payload); err != nil {
			return nil, err
		}
		if err = fw.Close(); err != nil {
			return nil, err
		}
		payload = buf.Bytes()
	}

	if maskingKey != nil {
		maskedPayload := make([]byte, len(payload))
		for i := range payload {
			maskedPayload[i] = payload[i] ^ maskingKey[i%4]
		}
		payload = maskedPayload
	}

	return payload, nil
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
