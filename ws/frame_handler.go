package ws

import (
	"bufio"
	"encoding/binary"
	"errors"
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

func (fr *frameHandler) Read(buf []byte) (int, error) {
	fr.rmu.Lock()
	defer fr.rmu.Unlock()

	if fr.dead {
		return -1, specs.ErrClosed
	}

	if fr.currentFrame == nil {
		frame, err := fr.pickFrame()
		if err != nil {
			return -1, err
		}
		if frame == nil {
			return -1, io.EOF
		}

		fr.currentFrame = frame
	}

	i, err := fr.currentFrame.Read(buf)
	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, specs.ErrClosed) {
		fr.currentFrame = nil
	}
	return i, err
}

func (fr *frameHandler) Write(msg []byte) (int, error) {
	return fr.writeFrame(wsBinaryFrame, msg)
}

func (fr *frameHandler) pickFrame() (*frameData, error) {
	frame, err := fr.readHeader()
	if err != nil {
		return nil, err
	}

	// The client MUST mask all frames sent to the server.
	if fr.isServer && frame.MaskingKey == nil {
		fr.WriteClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	// The server MUST NOT mask all frames.
	if !fr.isServer && frame.MaskingKey != nil {
		fr.WriteClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
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
		payload := make([]byte, maxControlPayload)
		n, _ := io.ReadFull(frame, payload[:min(int(frame.Length), maxControlPayload)])
		io.Copy(io.Discard, frame)
		if frame.Code == wsPingFrame {
			fr.writeFrame(wsPongFrame, payload[:n])
		}
		return nil, nil
	default:
		return nil, specs.ErrProtocol
	}

	return frame, nil
}

func (fr *frameHandler) readHeader() (*frameData, error) {
	first, err := fr.rws.ReadByte()
	if err != nil {
		return nil, err
	}

	var header frameData
	header.Fin = (first & 0x80) != 0
	header.Code = wsFrameType(first & 0x0F)

	maskAndLen, err := fr.rws.ReadByte()
	if err != nil {
		return nil, err
	}
	masked := (maskAndLen & 0x80) != 0
	length := int64(maskAndLen & 0x7F)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(fr.rws, ext[:]); err != nil {
			return nil, err
		}
		length = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(fr.rws, ext[:]); err != nil {
			return nil, err
		}
		length = int64(binary.BigEndian.Uint64(ext[:]))
	}

	if masked {
		header.MaskingKey = make([]byte, 4)
		if _, err = io.ReadFull(fr.rws, header.MaskingKey); err != nil {
			return nil, err
		}
	}

	header.Length = length
	header.reader = io.LimitReader(fr.rws, length)
	return &header, nil
}

func (fr *frameHandler) writeFrame(ft wsFrameType, payload []byte) (int, error) {
	fr.wmu.Lock()
	defer fr.wmu.Unlock()

	if fr.dead {
		return 0, specs.ErrClosed
	}

	if ft == wsCloseFrame {
		fr.dead = true
	}

	header := []byte{0x80 | byte(ft)}
	var maskKey []byte
	var err error

	if !fr.isServer {
		maskKey, err = newMask()
		if err != nil {
			return 0, err
		}
	}

	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, byte(length)|maskBit(maskKey))
	case length < 65536:
		header = append(header, 126|maskBit(maskKey))
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	default:
		header = append(header, 127|maskBit(maskKey))
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}

	if maskKey != nil {
		header = append(header, maskKey...)
	}

	if _, err = fr.rws.Write(header); err != nil {
		return 0, err
	}

	if maskKey != nil {
		maskedPayload := make([]byte, length)
		for i := range payload {
			maskedPayload[i] = payload[i] ^ maskKey[i%4]
		}
		if _, err = fr.rws.Write(maskedPayload); err != nil {
			return 0, err
		}
	} else {
		if _, err = fr.rws.Write(payload); err != nil {
			return 0, err
		}
	}

	return length, fr.rws.Flush()
}

func (fr *frameHandler) WriteClose(closeCode WsCloseCode) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(closeCode))
	_, err := fr.writeFrame(wsCloseFrame, buf)
	return err
}
