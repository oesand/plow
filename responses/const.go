package responses

import (
	"errors"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
)

type WebSocketHandler func(conn *WebSocketConn)

var bufioReaderPool internal.BufioReaderPool

var (
	websocketOp = specs.GigletOp("websocket")

	ErrorWebsocketInvalidFrameType = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("invalid frame type")}
	ErrorWebsocketFrameSizeExceed = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("frame size exceed")}
	ErrorWebsocketClosed = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("closed")}

	ErrorWebsocketNoRsV1 = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("rsv1 not implemented")}
	ErrorWebsocketNoRsV2 = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("rsv2 not implemented")}
	ErrorWebsocketNoRsV3 = &specs.GigletError{
		Op:  websocketOp,
		Err: errors.New("rsv3 not implemented")}
)

const (
	// Frame header byte 0 bits from Section 5.2 of RFC 6455
	websocketFinalBit byte = 1 << 7
	websocketRsv1Bit  byte = 1 << 6
	websocketRsv2Bit  byte = 1 << 5
	websocketRsv3Bit  byte = 1 << 4

	// Frame header byte 1 bits from Section 5.2 of RFC 6455
	websocketMaskBit byte = 1 << 7

	websocketMaxFrameHeaderSize         = 2 + 8 + 4 // Fixed header + length + mask
	websocketMaxServiceFramePayloadSize = 125

	// minCompressionLevel     = -2 // flate.HuffmanOnly not defined in Go < 1.6
	// maxCompressionLevel     = flate.BestCompression
	// defaultCompressionLevel = 1
)
