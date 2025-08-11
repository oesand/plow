package ws

// WsCloseCode represents the WebSocket close codes as defined in RFC 6455.
type WsCloseCode uint16

// Predefined WebSocket close codes as per RFC 6455.
// These codes are used to indicate the reason for closing a WebSocket connection.
const (
	CloseCodeNormal              WsCloseCode = 1000
	CloseCodeGoingAway           WsCloseCode = 1001
	CloseCodeProtocolError       WsCloseCode = 1002
	CloseCodeUnsupportedData     WsCloseCode = 1003
	CloseCodeNoStatusReceived    WsCloseCode = 1005
	CloseCodeAbnormal            WsCloseCode = 1006
	CloseCodeInvalidPayloadData  WsCloseCode = 1007
	CloseCodePolicyViolation     WsCloseCode = 1008
	CloseCodeMessageTooBig       WsCloseCode = 1009
	CloseCodeMandatoryExtension  WsCloseCode = 1010
	CloseCodeInternalServerError WsCloseCode = 1011
	CloseCodeServiceRestart      WsCloseCode = 1012
	CloseCodeTryAgainLater       WsCloseCode = 1013
	CloseCodeTLSHandshake        WsCloseCode = 1015
)
