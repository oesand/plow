package ws

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/internal/catch"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"sync"
	"time"
)

func newWsConn(
	ctx context.Context,
	conn net.Conn,
	isServer bool,
	compressEnabled bool,

	maxFrameSize int,
	readTimeout time.Duration,
	writeTimeout time.Duration,

	protocol string,
) *wsConn {
	reader := stream.DefaultBufioReaderPool.Get(conn)
	writer := stream.DefaultBufioWriterPool.Get(conn)
	rws := bufio.NewReadWriter(reader, writer)
	return &wsConn{
		ctx:  ctx,
		conn: conn,
		rws:  rws,

		isServer:        isServer,
		compressEnabled: compressEnabled,

		maxFrameSize: maxFrameSize,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,

		protocol: protocol,
	}
}

type wsConn struct {
	_ internal.NoCopy

	ctx  context.Context
	conn net.Conn
	rws  *bufio.ReadWriter

	isServer        bool
	compressEnabled bool

	maxFrameSize int
	readTimeout  time.Duration
	writeTimeout time.Duration

	protocol string

	closed bool
	dead   bool

	continuedFrame *frameHeader
	currentReader  io.Reader

	mu sync.Mutex
}

// Public functions. Ensure to call it with the mutex locked.

func (conn *wsConn) Alive() bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	return !conn.dead
}

func (conn *wsConn) Read(buf []byte) (int, error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return 0, specs.ErrClosed
	}

	err := conn.beforeOp()
	if err != nil {
		return 0, err
	}

	if conn.currentReader == nil {
		header, err := conn.readHeader()

		if err != nil {
			return 0, catch.CatchCommonErr(err)
		}
		if header == nil {
			return 0, io.EOF
		}

		if (conn.continuedFrame != nil && header.Type != wsContinuationFrame) ||
			(conn.continuedFrame == nil && header.Type == wsContinuationFrame) {
			conn.writeClose(CloseCodeProtocolError)
			return 0, specs.ErrProtocol
		}

		decompress := header.Rsv1Flag || (conn.continuedFrame != nil && conn.continuedFrame.Rsv1Flag)
		conn.currentReader = framePayloadReader(conn.rws.Reader, header.Length, header.MaskingKey, decompress)

		if !header.Fin {
			conn.continuedFrame = header
		} else {
			conn.continuedFrame = nil
		}
	}

	i, err := conn.currentReader.Read(buf)

	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, specs.ErrClosed) {
		conn.currentReader = nil
	}

	return i, catch.CatchCommonErr(err)
}

func (conn *wsConn) Write(payload []byte) (int, error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return 0, specs.ErrClosed
	}

	err := conn.beforeOp()
	if err != nil {
		return 0, err
	}

	return conn.writeFrame(wsBinaryFrame, payload)
}

func (conn *wsConn) WriteText(payload string) (int, error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return 0, specs.ErrClosed
	}

	err := conn.beforeOp()
	if err != nil {
		return 0, err
	}

	return conn.writeFrame(wsTextFrame, []byte(payload))
}

func (conn *wsConn) WriteClose(closeCode WsCloseCode) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return specs.ErrClosed
	}

	err := conn.beforeOp()
	if err != nil {
		return err
	}

	return conn.writeClose(closeCode)
}

func (conn *wsConn) RemoteAddr() net.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *wsConn) Protocol() string {
	return conn.protocol
}

func (conn *wsConn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.closed {
		return specs.ErrClosed
	}

	conn.closed = true
	conn.dead = true

	stream.DefaultBufioReaderPool.Put(conn.rws.Reader)
	stream.DefaultBufioWriterPool.Put(conn.rws.Writer)

	return conn.conn.Close()
}

// Private functions. Ensure to call it without the mutex locked.

func (conn *wsConn) beforeOp() error {
	if conn.readTimeout > 0 {
		err := conn.conn.SetReadDeadline(time.Now().Add(conn.readTimeout))
		if err != nil {
			return err
		}
	}

	if conn.writeTimeout > 0 {
		err := conn.conn.SetWriteDeadline(time.Now().Add(conn.writeTimeout))
		if err != nil {
			return err
		}
	}

	if err := catch.CatchContextCancel(conn.ctx); err != nil {
		return err
	}

	return nil
}

func (conn *wsConn) readHeader() (*frameHeader, error) {
	header, err := readFrameHeader(conn.rws.Reader)
	if err != nil {
		return nil, err
	}

	// Маска по сторонам
	if conn.isServer && header.MaskingKey == nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}
	if !conn.isServer && header.MaskingKey != nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	// Ограничение длины кадра
	if conn.maxFrameSize > 0 && header.Length > conn.maxFrameSize {
		conn.writeClose(CloseCodeMessageTooBig)
		return nil, specs.ErrTooLarge
	}

	switch header.Type {
	case wsContinuationFrame:
		if header.Rsv1Flag || header.Rsv2Flag || header.Rsv3Flag {
			conn.writeClose(CloseCodeProtocolError)
			return nil, specs.ErrProtocol
		}
	case wsTextFrame, wsBinaryFrame:
		if !conn.compressEnabled && (header.Rsv1Flag || header.Rsv2Flag || header.Rsv3Flag) {
			conn.writeClose(CloseCodeProtocolError)
			return nil, specs.ErrProtocol
		}
	case wsCloseFrame:
		conn.dead = true
		if !header.Fin || header.Length > maxControlPayload || header.Rsv1Flag || header.Rsv2Flag || header.Rsv3Flag {
			return nil, specs.ErrProtocol
		}
		io.CopyN(io.Discard, conn.rws.Reader, int64(header.Length))
		return nil, specs.ErrClosed
	case wsPingFrame, wsPongFrame:
		if !header.Fin || header.Length > maxControlPayload || header.Rsv1Flag || header.Rsv2Flag || header.Rsv3Flag {
			conn.writeClose(CloseCodeProtocolError)
			return nil, specs.ErrProtocol
		}
		if header.Type == wsPingFrame {
			payload := make([]byte, header.Length)
			reader := framePayloadReader(conn.rws.Reader, header.Length, header.MaskingKey, false)
			_, err = io.ReadFull(reader, payload)
			if err != nil {
				return nil, err
			}
			conn.writeFrameLowLevel(wsPongFrame, payload, true, false)
		}
		return nil, nil
	default:
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	return header, nil
}

// Private functions for writing frames.

func (conn *wsConn) writeFrame(frameType wsFrameType, payload []byte) (int, error) {
	if frameType != wsTextFrame && frameType != wsBinaryFrame {
		return 0, specs.ErrProtocol
	}

	if len(payload) == 0 {
		return 0, nil
	}

	var err error
	if conn.compressEnabled {
		payload, err = compressFramePayload(payload)
		if err != nil {
			return 0, err
		}
	}

	maxSize := conn.maxFrameSize
	if maxSize <= 0 {
		maxSize = len(payload)
	}

	total := 0
	first := true
	for offset := 0; offset < len(payload); {
		chunkEnd := min(len(payload), offset+maxSize)
		chunk := payload[offset:chunkEnd]
		offset = chunkEnd

		final := offset >= len(payload)
		rsv1 := conn.compressEnabled && first

		ft := frameType
		if !first {
			ft = wsContinuationFrame
		}

		n, err := conn.writeFrameLowLevel(ft, chunk, final, rsv1)
		if err != nil {
			return total, catch.CatchCommonErr(err)
		}
		total += n
		first = false
	}

	return total, nil
}

func (conn *wsConn) writeClose(closeCode WsCloseCode) error {
	buf := binary.BigEndian.AppendUint16(nil, uint16(closeCode))
	_, err := conn.writeFrameLowLevel(wsCloseFrame, buf, true, false)
	return err
}

func (conn *wsConn) writeFrameLowLevel(ft wsFrameType, payload []byte, final bool, rsv1 bool) (int, error) {
	var maskingKey []byte
	if !conn.isServer {
		maskingKey = make([]byte, 4)
		if _, err := io.ReadFull(rand.Reader, maskingKey); err != nil {
			return 0, err
		}
	}

	header := &frameHeader{
		Fin:        final,
		Type:       ft,
		Rsv1Flag:   rsv1,
		Rsv2Flag:   false,
		Rsv3Flag:   false,
		Length:     len(payload),
		MaskingKey: maskingKey,
	}

	if ft >= wsCloseFrame {
		if !final {
			return 0, specs.ErrProtocol
		}
		if header.Length > 125 {
			return 0, specs.ErrProtocol
		}
		if rsv1 {
			return 0, specs.ErrProtocol
		}
	}

	preparedHeader := prepareFrameHeader(header)
	if _, err := conn.rws.Write(preparedHeader); err != nil {
		return 0, err
	}

	if maskingKey != nil {
		maskedPayload := make([]byte, len(payload))
		for i := range payload {
			maskedPayload[i] = payload[i] ^ maskingKey[i%4]
		}
		payload = maskedPayload
	}

	if _, err := conn.rws.Write(payload); err != nil {
		return 0, err
	}
	if err := conn.rws.Flush(); err != nil {
		return 0, err
	}

	return header.Length, nil
}
