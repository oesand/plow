package ws

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"slices"
	"sync"
	"time"
)

func newWsConn(
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

	err := conn.applyTimeout()
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

func (conn *wsConn) Write(msg []byte) (int, error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return 0, specs.ErrClosed
	}

	err := conn.applyTimeout()
	if err != nil {
		return 0, err
	}

	if maxSize := conn.maxFrameSize; maxSize > 10 && len(msg) > maxSize {
		// TODO : optimize this to avoid unnecessary allocations
		chunks := slices.Collect(slices.Chunk(msg, maxSize))
		var length int
		for i, chunk := range chunks {
			if i == 0 {
				length, err = conn.writeFrame(wsBinaryFrame, chunk, false)
				if err != nil {
					return 0, catch.CatchCommonErr(err)
				}
				continue
			}

			n, err := conn.writeFrame(wsContinuationFrame, chunk, i == len(chunks)-1)
			if err != nil {
				return 0, catch.CatchCommonErr(err)
			}
			length += n
		}

		return length, nil
	}

	n, err := conn.writeFrame(wsBinaryFrame, msg, true)
	return n, catch.CatchCommonErr(err)
}

func (conn *wsConn) WriteClose(closeCode WsCloseCode) error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.dead {
		return specs.ErrClosed
	}

	err := conn.applyTimeout()
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

func (conn *wsConn) applyTimeout() error {
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

	return nil
}

func (conn *wsConn) readHeader() (*frameHeader, error) {
	header, err := readFrameHeader(conn.rws.Reader)
	if err != nil {
		return nil, err
	}

	// The client MUST mask all frames sent to the server.
	if conn.isServer && header.MaskingKey == nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	// The server MUST NOT mask all frames.
	if !conn.isServer && header.MaskingKey != nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

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
		break
	case wsTextFrame, wsBinaryFrame:
		break
	case wsCloseFrame:
		conn.dead = true
		return nil, specs.ErrClosed
	case wsPingFrame, wsPongFrame:
		if !header.Fin || header.Rsv1Flag || header.Rsv2Flag || header.Rsv3Flag {
			conn.writeClose(CloseCodeProtocolError)
			return nil, specs.ErrProtocol
		}
		if header.Type == wsPingFrame {
			maxlen := min(header.Length, maxControlPayload)
			payload := make([]byte, maxlen)
			reader := framePayloadReader(conn.rws.Reader, maxlen, header.MaskingKey, false)
			n, _ := io.ReadFull(reader, payload)
			conn.writeFrame(wsPongFrame, payload[:n], true)
		}
		io.Copy(io.Discard, conn.rws.Reader)
		return nil, nil
	default:
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	return header, nil
}

func (conn *wsConn) writeFrame(ft wsFrameType, payload []byte, final bool) (int, error) {
	var maskingKey []byte
	var err error

	if !conn.isServer {
		maskingKey = make([]byte, 4)
		if _, err = io.ReadFull(rand.Reader, maskingKey); err != nil {
			return 0, err
		}
	}

	mustCompress := conn.compressEnabled && (ft == wsBinaryFrame || ft == wsTextFrame || ft == wsContinuationFrame)
	preparedPayload, err := prepareFramePayload(maskingKey, mustCompress, payload)
	if err != nil {
		return 0, err
	}

	header := &frameHeader{
		Fin:        final,
		Type:       ft,
		Rsv1Flag:   mustCompress && ft != wsContinuationFrame,
		Length:     len(preparedPayload),
		MaskingKey: maskingKey,
	}

	preparedHeader := prepareFrameHeader(header)
	_, err = conn.rws.Write(preparedHeader)
	if err != nil {
		return 0, err
	}

	if _, err = conn.rws.Write(preparedPayload); err != nil {
		return 0, err
	}

	err = conn.rws.Flush()
	if err != nil {
		return 0, err
	}

	return header.Length, nil
}

func (conn *wsConn) writeClose(closeCode WsCloseCode) error {
	buf := binary.BigEndian.AppendUint16(nil, uint16(closeCode))
	_, err := conn.writeFrame(wsCloseFrame, buf, true)
	return err
}
