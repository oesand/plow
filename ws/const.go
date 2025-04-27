package ws

import (
	"crypto/sha1"
	"encoding/base64"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
)

type WebSocketHandler func(conn *WebSocketConn)

var bufioReaderPool utils.BufioReaderPool

var (
	websocketAcceptBaseKey = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

	websocketOp = specs.GigletOp("websocket")

	ErrorWebsocketInvalidFrameType = specs.NewOpError(websocketOp, "invalid frame type")
	ErrorWebsocketFrameSizeExceed  = specs.NewOpError(websocketOp, "frame size exceed")
	ErrorWebsocketClosed           = specs.NewOpError(websocketOp, "closed")

	ErrorWebsocketNoRsV1 = specs.NewOpError(websocketOp, "rsv1 not implemented")
	ErrorWebsocketNoRsV2 = specs.NewOpError(websocketOp, "rsv2 not implemented")
	ErrorWebsocketNoRsV3 = specs.NewOpError(websocketOp, "rsv3 not implemented")
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

func ComputeWebSocketAcceptKey(challengeKey string) string {
	h := sha1.New() // (CWE-326) -- https://datatracker.ietf.org/doc/html/rfc6455#page-54
	h.Write([]byte(challengeKey))
	h.Write(websocketAcceptBaseKey)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
