package ws

import (
	"bufio"
	"encoding/binary"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"sync"
)

func newFrameHandler(rws *bufio.ReadWriter, isServer bool) *frameHandler {
	return &frameHandler{
		rws:      rws,
		isServer: isServer,
	}
}

type frameHandler struct {
	_ internal.NoCopy

	rws      *bufio.ReadWriter
	isServer bool

	dead         bool
	currentFrame *frameData
	rmu          sync.Mutex
	wmu          sync.Mutex
}

func (fr *frameHandler) Alive() bool {
	return !fr.dead
}

func (fr *frameHandler) Read(msg []byte) (int, error) {
	fr.rmu.Lock()
	defer fr.rmu.Unlock()

	if fr.dead {
		return 0, specs.ErrClosed
	}

	var frame *frameData
	if fr.currentFrame != nil {
		frame = fr.currentFrame
	} else {
		var err error
		frame, err = fr.pickFrame()
		if err != nil {
			return -1, err
		}
		fr.currentFrame = frame
	}

	i, err := frame.Read(msg)
	if err == io.EOF || err == specs.ErrClosed {
		fr.currentFrame = nil
	}
	return i, err
}

func (fr *frameHandler) Write(msg []byte) error {
	return fr.writeFrame(wsBinaryFrame, msg)
}

func (fr *frameHandler) pickFrame() (*frameData, error) {
	frame, err := fr.readHeader()
	if err != nil {
		return nil, err
	}

	if fr.isServer {
		// The client MUST mask all frames sent to the server.
		if frame.MaskingKey == nil {
			fr.WriteClose(CloseCodeProtocolError)
			return nil, specs.ErrClosed
		}
	} else {
		// The server MUST NOT mask all frames.
		if frame.MaskingKey != nil {
			fr.WriteClose(CloseCodeProtocolError)
			return nil, specs.ErrClosed
		}
	}

	switch frame.Code {
	case wsContinuationFrame:
		frame.Code = wsBinaryFrame
	case wsTextFrame, wsBinaryFrame:
		break
	case wsCloseFrame:
		fr.dead = true
		return nil, specs.ErrClosed
	case wsPingFrame, wsPongFrame:
		b := make([]byte, maxServiceFramePayloadSize)
		n, err := io.ReadFull(frame, b)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		io.Copy(io.Discard, frame)
		if frame.Code == wsPingFrame {
			if err = fr.writeFrame(wsPongFrame, b[:n]); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}

	fr.currentFrame = frame
	return frame, nil
}

func (fr *frameHandler) readHeader() (*frameData, error) {
	// First byte. FIN/RSV1/RSV2/RSV3/Code(4bits)
	b, err := fr.rws.ReadByte()
	if err != nil {
		return nil, err
	}

	var header frameData
	header.Fin = ((b >> 7) & 1) != 0
	header.Code = WsFrameType(b & 0x0f)

	for i := 0; i < 3; i++ {
		j := uint(6 - i)
		header.Rsv[i] = ((b >> j) & 1) != 0
	}

	var data []byte
	data = append(data, b)

	// Second byte. Mask/Payload len(7bits)
	b, err = fr.rws.ReadByte()
	if err != nil {
		return nil, err
	}

	mask := (b & 0x80) != 0
	b &= 0x7f
	lengthFields := 0
	switch {
	case b <= 125: // Payload length 7bits.
		header.Length = int64(b)
	case b == 126: // Payload length 7+16bits
		lengthFields = 2
	case b == 127: // Payload length 7+64bits
		lengthFields = 8
	}

	data = append(data, b)

	for i := 0; i < lengthFields; i++ {
		b, err = fr.rws.ReadByte()
		if err != nil {
			return nil, err
		}
		if lengthFields == 8 && i == 0 { // MSB must be zero when 7+64 bits
			b &= 0x7f
		}
		data = append(data, b)
		header.Length = header.Length*256 + int64(b)
	}
	if mask {
		// Masking key. 4 bytes.
		for i := 0; i < 4; i++ {
			b, err = fr.rws.ReadByte()
			if err != nil {
				return nil, err
			}
			data = append(data, b)
			header.MaskingKey = append(header.MaskingKey, b)
		}
	}

	header.reader = io.LimitReader(fr.rws, header.Length)
	header.length = len(data) + int(header.Length)
	return &header, nil
}

func (fr *frameHandler) writeFrame(frameType WsFrameType, payload []byte) (err error) {
	fr.wmu.Lock()
	defer fr.wmu.Unlock()

	if fr.dead {
		return specs.ErrClosed
	}

	var maskingKey []byte
	if !fr.isServer {
		maskingKey, err = newFrameMask()
		if err != nil {
			return err
		}
	}

	if frameType == wsCloseFrame {
		fr.dead = true
	}

	var header []byte
	var b byte
	b |= 0x80
	b |= byte(frameType)

	header = append(header, b)

	if maskingKey != nil {
		b = 0x80
	} else {
		b = 0
	}
	lengthFields := 0
	length := len(payload)
	switch {
	case length <= 125:
		b |= byte(length)
	case length < 65536:
		b |= 126
		lengthFields = 2
	default:
		b |= 127
		lengthFields = 8
	}
	header = append(header, b)
	for i := 0; i < lengthFields; i++ {
		j := uint((lengthFields - i - 1) * 8)
		b = byte((length >> j) & 0xff)
		header = append(header, b)
	}
	if maskingKey != nil {
		header = append(header, maskingKey...)
		fr.rws.Write(header)
		data := make([]byte, length)
		for i := range data {
			data[i] = payload[i] ^ maskingKey[i%4]
		}
		fr.rws.Write(data)
		return fr.rws.Flush()
	}
	fr.rws.Write(header)
	fr.rws.Write(payload)
	return fr.rws.Flush()
}

func (fr *frameHandler) WriteClose(closeCode WsCloseCode) error {
	if fr.dead {
		return specs.ErrClosed
	}

	msg := make([]byte, 2)
	binary.BigEndian.PutUint16(msg, uint16(closeCode))
	return fr.writeFrame(wsCloseFrame, msg)
}
