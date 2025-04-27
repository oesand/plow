package ws

type WebSocketFrame uint16

const (
	WebSocketUnknownFrame WebSocketFrame = 0
	WebSocketTextFrame    WebSocketFrame = 1
	WebSocketBinaryFrame  WebSocketFrame = 2
	WebSocketCloseFrame   WebSocketFrame = 8
	WebSocketPingFrame    WebSocketFrame = 9
	WebSocketPongFrame    WebSocketFrame = 10
)

func (frame WebSocketFrame) IsService() bool {
	return frame == WebSocketCloseFrame || frame == WebSocketPingFrame || frame == WebSocketPongFrame
}

func (frame WebSocketFrame) IsContent() bool {
	return frame == WebSocketTextFrame || frame == WebSocketBinaryFrame
}
