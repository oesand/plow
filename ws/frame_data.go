package ws

import "io"

type frameData struct {
	Fin        bool
	Rsv        [3]bool
	Code       WsFrameType
	Length     int64
	MaskingKey []byte

	reader io.Reader
	pos    int64
	length int
}

func (frame *frameData) Read(msg []byte) (n int, err error) {
	n, err = frame.reader.Read(msg)
	if frame.MaskingKey != nil {
		for i := 0; i < n; i++ {
			msg[i] = msg[i] ^ frame.MaskingKey[frame.pos%4]
			frame.pos++
		}
	}
	return n, err
}
