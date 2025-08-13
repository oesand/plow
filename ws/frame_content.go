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

func compressFramePayload(data []byte) ([]byte, error) {
	// Compress the payload using per-message deflate
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.BestSpeed)
	if err != nil {
		return nil, err
	}
	if _, err = w.Write(data); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}

	// Remove the deflate tail if it exists
	payload := buf.Bytes()
	if len(payload) >= 4 {
		n := len(payload)
		if payload[n-4] == 0x00 && payload[n-3] == 0x00 && payload[n-2] == 0xFF && payload[n-1] == 0xFF {
			return payload[:n-4], nil
		}
	}
	return payload, nil
}
