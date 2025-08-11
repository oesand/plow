package ws

import "io"

type frameData struct {
	Fin        bool
	Code       wsFrameType
	Length     int64
	MaskingKey []byte

	reader io.Reader
	pos    int64
}

func (frame *frameData) Read(p []byte) (int, error) {
	n, err := frame.reader.Read(p)
	if frame.MaskingKey != nil {
		for i := 0; i < n; i++ {
			p[i] ^= frame.MaskingKey[frame.pos%4]
			frame.pos++
		}
	}
	return n, err
}
