package ws

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

type Conf struct {
	EnableCompression bool
	ReadLimit         int64
}

type wsConn struct {
	request giglet.Request
	conn    net.Conn
	reader  bufio.Reader
	conf    Conf
	dead    bool
}

func (conn *wsConn) Context() context.Context {
	return conn.request.Context()
}

func (conn *wsConn) WithContext(context context.Context) {
	conn.request.WithContext(context)
}

func (conn *wsConn) RemoteAddr() net.Addr {
	return conn.request.RemoteAddr()
}

func (conn *wsConn) Url() *specs.Url {
	return conn.request.Url()
}

func (conn *wsConn) Header() *specs.Header {
	return conn.request.Header()
}

func (conn *wsConn) Alive() bool {
	return !conn.dead
}

func (conn *wsConn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *wsConn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *wsConn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

func (conn *wsConn) Read() (frame WebSocketFrame, buf []byte, err error) {
	if conn.dead {
		frame, err = WebSocketCloseFrame, ErrorWebsocketClosed
		return
	}

	buf, err = conn.peekReader(2)
	if err != nil {
		return
	}

	frame = WebSocketFrame(buf[0] & 0xf)
	final := buf[0]&websocketFinalBit != 0
	//rsv1 := buf[0]&websocketRsv1Bit != 0
	//rsv2 := buf[0]&websocketRsv2Bit != 0
	//rsv3 := buf[0]&websocketRsv3Bit != 0
	// mask := buf[1] & websocketMaskBit != 0

	remaining := int64(buf[1] & 0x7f)

	//if rsv1 {
	//	if conn.decompressionReader == nil {
	//		err = ErrorWebsocketNoRsV1
	//		return
	//	}
	//} else if rsv2 {
	//	conn.WriteCloseFrame(WebSocketCloseUnsupportedData)
	//	err = ErrorWebsocketNoRsV2
	//	return
	//} else if rsv3 {
	//	conn.WriteCloseFrame(WebSocketCloseUnsupportedData)
	//	err = ErrorWebsocketNoRsV3
	//	return
	//}

	if frame.IsService() {
		if remaining > websocketMaxServiceFramePayloadSize || !final {
			if frame != WebSocketCloseFrame {
				conn.Close(WebSocketCloseMessageTooBig)
			}
			err = ErrorWebsocketFrameSizeExceed
		}
		return
	} else if !frame.IsContent() {
		conn.Close(WebSocketCloseInvalidPayloadData)
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
		conn.Close(WebSocketCloseMessageTooBig)
		frame, err = WebSocketCloseFrame, ErrorWebsocketClosed
		return
	}

	//if frame.IsContent() && conn.decompressionReader != nil {
	//	var source io.Reader = conn.conn
	//	if conn.conf.ReadLimit > 0 {
	//		source = io.LimitReader(source, conn.conf.ReadLimit)
	//	}
	//	reader := conn.decompressionReader(source)
	//	_, err = reader.Read(buf[:])
	//	reader.Close()
	//} else {
	//	buf, err = conn.peekReader(int(remaining))
	//}
	return
}

func (conn *wsConn) Write(frameType WebSocketFrame, payload []byte) (err error) {
	if conn.dead {
		return ErrorWebsocketClosed
	} else if !frameType.IsService() && !frameType.IsContent() {
		return ErrorWebsocketInvalidFrameType
	} else if frameType.IsService() && len(payload) > websocketMaxServiceFramePayloadSize {
		return ErrorWebsocketFrameSizeExceed
	}

	frameByte := byte(frameType)
	if frameType == WebSocketCloseFrame {
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

	//if frameType.IsContent() && conn.compressionWriter != nil {
	//	writer := conn.compressionWriter(conn.conn)
	//	_, err = writer.WriteTo(buf)
	//	writer.Close()
	//} else {
	//	_, err = conn.conn.WriteTo(buf)
	//}

	if err == io.EOF {
		conn.dead = true
	} else if frameType == WebSocketCloseFrame {
		conn.dead = true
		err = io.EOF
		conn.conn.Close()
	}
	return err
}

func (conn *wsConn) Close(reason WebSocketClose) error {
	if conn.dead {
		return ErrorWebsocketClosed
	}

	payload := strconv.AppendUint(nil, uint64(reason), 10)
	payload = append(payload, ' ')
	payload = append(payload, reason.Detail()...)

	return conn.Write(WebSocketCloseFrame, payload)
}

func (conn *wsConn) peekReader(n int) (buf []byte, err error) {
	buf, err = conn.reader.Peek(n)
	if err == io.EOF {
		err = ErrorWebsocketClosed
		conn.dead = true
		return
	}
	conn.reader.Discard(n)
	return
}
