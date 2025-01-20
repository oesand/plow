package responses

import (
	"bufio"
	"context"
	"encoding/binary"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"strconv"
	"time"
)

type WebSocketConf struct {
	EnableCompression bool
	ReadLimit         int64
}

type WebSocketConn struct {
	request giglet.Request
	conn    net.Conn
	reader  bufio.Reader
	conf    WebSocketConf
	dead    bool

	compressionWriter   func(io.WriteCloser) io.WriteCloser
	decompressionReader func(io.Reader) io.ReadCloser
}

func (conn *WebSocketConn) Context() any {
	return conn.request.Context()
}

func (conn *WebSocketConn) PutContext(context context.Context) {
	conn.request.PutContext(context)
}

func (conn *WebSocketConn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *WebSocketConn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *WebSocketConn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

func (conn *WebSocketConn) RemoteAddr() net.Addr {
	return conn.request.RemoteAddr()
}

func (conn *WebSocketConn) Url() *specs.Url {
	return conn.request.Url()
}

func (conn *WebSocketConn) Header() *specs.ReadOnlyHeader {
	return conn.request.Header()
}

func (conn *WebSocketConn) Alive() bool {
	return !conn.dead
}

func (conn *WebSocketConn) WriteFrame(frameType specs.WebSocketFrame, payload []byte) (err error) {
	if conn.dead {
		return ErrorWebsocketClosed
	} else if !frameType.IsService() && !frameType.IsContent() {
		return ErrorWebsocketInvalidFrameType
	} else if frameType.IsService() && len(payload) > websocketMaxServiceFramePayloadSize {
		return ErrorWebsocketFrameSizeExceed
	}

	frameByte := byte(frameType)
	if frameType == specs.WebSocketCloseFrame {
		frameByte |= websocketFinalBit
	}
	// if w.compress {
	// 	b0 |= rsv1Bit
	// }

	length := len(payload)
	buf := make([]byte, 0, websocketMaxFrameHeaderSize+length) // max header & frame size
	buf = append(buf, frameByte)
	if length >= 65536 {
		buf = append(buf, 127)
	} else if length > 125 {
		buf = append(buf, 126)
	}
	buf = append(buf, byte(length))
	buf = append(buf, payload...)

	if frameType.IsContent() && conn.compressionWriter != nil {
		writer := conn.compressionWriter(conn.conn)
		_, err = writer.Write(buf)
		writer.Close()
	} else {
		_, err = conn.conn.Write(buf)
	}

	if err == io.EOF {
		conn.dead = true
	} else if frameType == specs.WebSocketCloseFrame {
		conn.dead = true
		err = io.EOF
		conn.conn.Close()
	}
	return err
}

func (conn *WebSocketConn) WriteCloseFrame(reason specs.WebSocketClose) error {
	if conn.dead {
		return ErrorWebsocketClosed
	}

	payload := strconv.AppendUint(nil, uint64(reason), 10)
	payload = append(payload, ' ')
	payload = append(payload, reason.Detail()...)

	return conn.WriteFrame(specs.WebSocketCloseFrame, payload)
}

func (conn *WebSocketConn) peekReader(n int) (buf []byte, err error) {
	buf, err = conn.reader.Peek(n)
	if err == io.EOF {
		err = ErrorWebsocketClosed
		conn.dead = true
		return
	}
	conn.reader.Discard(n)
	return
}

func (conn *WebSocketConn) ReadFrame() (frame specs.WebSocketFrame, buf []byte, err error) {
	if conn.dead {
		frame, err = specs.WebSocketCloseFrame, ErrorWebsocketClosed
		return
	}

	buf, err = conn.peekReader(2)
	if err != nil {
		return
	}

	frame = specs.WebSocketFrame(buf[0] & 0xf)
	final := buf[0]&websocketFinalBit != 0
	rsv1 := buf[0]&websocketRsv1Bit != 0
	rsv2 := buf[0]&websocketRsv2Bit != 0
	rsv3 := buf[0]&websocketRsv3Bit != 0
	// mask := buf[1] & websocketMaskBit != 0

	remaining := int64(buf[1] & 0x7f)

	if rsv1 {
		if conn.decompressionReader == nil {
			err = ErrorWebsocketNoRsV1
			return
		}
	} else if rsv2 {
		conn.WriteCloseFrame(specs.WebSocketCloseUnsupportedData)
		err = ErrorWebsocketNoRsV2
		return
	} else if rsv3 {
		conn.WriteCloseFrame(specs.WebSocketCloseUnsupportedData)
		err = ErrorWebsocketNoRsV3
		return
	}

	if frame.IsService() {
		if remaining > websocketMaxServiceFramePayloadSize || !final {
			if frame != specs.WebSocketCloseFrame {
				conn.WriteCloseFrame(specs.WebSocketCloseMessageTooBig)
			}
			err = ErrorWebsocketFrameSizeExceed
		}
		return
	} else if !frame.IsContent() {
		conn.WriteCloseFrame(specs.WebSocketCloseInvalidPayloadData)
		err = ErrorWebsocketInvalidFrameType
		return
	}

	// Next handle only Content Frames...

	switch remaining {
	case 126:
		buf, err = conn.peekReader(2)
		if err != nil {
			return
		}
		remaining = int64(binary.BigEndian.Uint16(buf))

	case 127:
		buf, err = conn.peekReader(8)
		if err != nil {
			return
		}
		remaining = int64(binary.BigEndian.Uint32(buf))
	}

	if conn.conf.ReadLimit > 0 && remaining > conn.conf.ReadLimit {
		conn.WriteCloseFrame(specs.WebSocketCloseMessageTooBig)
		frame, err = specs.WebSocketCloseFrame, ErrorWebsocketClosed
		return
	}

	if frame.IsContent() && conn.decompressionReader != nil {
		var source io.Reader = conn.conn
		if conn.conf.ReadLimit > 0 {
			source = io.LimitReader(source, conn.conf.ReadLimit)
		}
		reader := conn.decompressionReader(source)
		_, err = reader.Read(buf[:])
		reader.Close()
	} else {
		buf, err = conn.peekReader(int(remaining))
	}
	return
}
