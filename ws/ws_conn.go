package ws

import (
	"bufio"
	"bytes"
	"compress/flate"
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
	compress bool,

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

		isServer: isServer,
		compress: compress,

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

	isServer bool
	compress bool

	maxFrameSize int
	readTimeout  time.Duration
	writeTimeout time.Duration

	protocol string

	closed        bool
	dead          bool
	currentFrame  *frameData
	continueFrame bool

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
		return -1, specs.ErrClosed
	}

	err := conn.applyTimeout()
	if err != nil {
		return -1, err
	}

	if conn.currentFrame == nil {
		frame, err := conn.pickFrame()
		if err != nil {
			return -1, err
		}
		if frame == nil {
			return -1, io.EOF
		}
		if (conn.continueFrame && frame.Code != wsContinuationFrame) ||
			(!conn.continueFrame && frame.Code == wsContinuationFrame) {
			conn.writeClose(CloseCodeProtocolError)
			return -1, specs.ErrProtocol
		}

		conn.continueFrame = !frame.Fin
		conn.currentFrame = frame
	}

	i, err := conn.currentFrame.Read(buf)
	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, specs.ErrClosed) {
		conn.currentFrame = nil
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
					return 0, err
				}
				continue
			}

			n, err := conn.writeFrame(wsContinuationFrame, chunk, i == len(chunks)-1)
			if err != nil {
				return 0, err
			}
			length += n
		}

		return length, nil
	}
	return conn.writeFrame(wsBinaryFrame, msg, true)
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

func (conn *wsConn) pickFrame() (*frameData, error) {
	frame, err := conn.readHeader()
	if err != nil {
		return nil, err
	}

	switch frame.Code {
	case wsContinuationFrame, wsTextFrame, wsBinaryFrame:
		break
	case wsCloseFrame:
		conn.dead = true
		return nil, specs.ErrClosed
	case wsPingFrame, wsPongFrame:
		if !frame.Fin {
			conn.writeClose(CloseCodeProtocolError)
			return nil, specs.ErrProtocol
		}
		if frame.Code == wsPingFrame {
			payload := make([]byte, min(frame.Length, maxControlPayload))
			n, _ := io.ReadFull(frame, payload)
			conn.writeFrame(wsPongFrame, payload[:n], true)
		}
		io.Copy(io.Discard, frame)
		return nil, nil
	default:
		return nil, specs.ErrProtocol
	}

	return frame, nil
}

func (conn *wsConn) readHeader() (*frameData, error) {
	first, err := conn.rws.ReadByte()
	if err != nil {
		return nil, err
	}

	var header frameData
	header.Fin = (first & 0x80) != 0
	header.Code = wsFrameType(first & 0x0F)

	maskAndLen, err := conn.rws.ReadByte()
	if err != nil {
		return nil, err
	}
	masked := (maskAndLen & 0x80) != 0
	length := int(maskAndLen & 0x7F)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(conn.rws, ext[:]); err != nil {
			return nil, err
		}
		length = int(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(conn.rws, ext[:]); err != nil {
			return nil, err
		}
		length = int(binary.BigEndian.Uint64(ext[:]))
	}

	if conn.maxFrameSize > 0 && length > conn.maxFrameSize {
		conn.writeClose(CloseCodeMessageTooBig)
		return nil, specs.ErrTooLarge
	}

	var maskingKey []byte
	if masked {
		maskingKey = make([]byte, 4)
		if _, err = io.ReadFull(conn.rws, maskingKey); err != nil {
			return nil, err
		}
	}

	// The client MUST mask all frames sent to the server.
	if conn.isServer && maskingKey == nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	// The server MUST NOT mask all frames.
	if !conn.isServer && maskingKey != nil {
		conn.writeClose(CloseCodeProtocolError)
		return nil, specs.ErrProtocol
	}

	header.Length = length
	reader := unmaskReader(conn.rws, int64(length), maskingKey)

	if conn.compress {
		reader = flate.NewReader(reader)
	}

	header.Reader = reader
	return &header, nil
}

func (conn *wsConn) writeFrame(ft wsFrameType, payload []byte, final bool) (int, error) {
	if conn.dead {
		return 0, specs.ErrClosed
	}

	if ft == wsCloseFrame {
		conn.dead = true
	}

	first := byte(ft)
	if final {
		first |= 0x80
	}
	if conn.compress {
		first |= 1 << 6 // Set RSV1 bit for compression
	}

	header := []byte{first}

	var maskKey []byte
	var maskBit byte
	var err error

	if !conn.isServer {
		maskKey, err = newMask()
		if err != nil {
			return 0, err
		}
		maskBit = 0x80
	}

	if conn.compress {
		var buf bytes.Buffer
		fw, err := flate.NewWriter(&buf, flate.BestSpeed)
		if err != nil {
			return 0, err
		}
		if _, err = fw.Write(payload); err != nil {
			return 0, err
		}
		if err = fw.Close(); err != nil {
			return 0, err
		}
		payload = buf.Bytes()
	}

	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, byte(length)|maskBit)
	case length < 65536:
		header = append(header, 126|maskBit)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	default:
		header = append(header, 127|maskBit)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}

	if maskKey != nil {
		header = append(header, maskKey...)
	}

	if _, err = conn.rws.Write(header); err != nil {
		return 0, err
	}

	if maskKey != nil {
		maskedPayload := make([]byte, length)
		for i := range payload {
			maskedPayload[i] = payload[i] ^ maskKey[i%4]
		}
		if _, err = conn.rws.Write(maskedPayload); err != nil {
			return 0, err
		}
	} else {
		if _, err = conn.rws.Write(payload); err != nil {
			return 0, err
		}
	}

	return length, conn.rws.Flush()
}

func (conn *wsConn) writeClose(closeCode WsCloseCode) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(closeCode))
	_, err := conn.writeFrame(wsCloseFrame, buf, true)
	return err
}
