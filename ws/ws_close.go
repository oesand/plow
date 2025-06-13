package ws

type WsCloseCode uint16

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
