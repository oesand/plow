package specs

type WebSocketClose uint16

const (
	WebSocketCloseNormal              WebSocketClose = 1000
	WebSocketCloseGoingAway           WebSocketClose = 1001
	WebSocketCloseProtocolError       WebSocketClose = 1002
	WebSocketCloseUnsupportedData     WebSocketClose = 1003
	WebSocketCloseNoStatusReceived    WebSocketClose = 1005
	WebSocketCloseAbnormal            WebSocketClose = 1006
	WebSocketCloseInvalidPayloadData  WebSocketClose = 1007
	WebSocketClosePolicyViolation     WebSocketClose = 1008
	WebSocketCloseMessageTooBig       WebSocketClose = 1009
	WebSocketCloseMandatoryExtension  WebSocketClose = 1010
	WebSocketCloseInternalServerError WebSocketClose = 1011
	WebSocketCloseServiceRestart      WebSocketClose = 1012
	WebSocketCloseTryAgainLater       WebSocketClose = 1013
	WebSocketCloseTLSHandshake        WebSocketClose = 1015
)

func (code WebSocketClose) Detail() []byte {
	switch code {
	case WebSocketCloseNormal:
		return []byte("(normal)")
	case WebSocketCloseGoingAway:
		return []byte("(going away)")
	case WebSocketCloseProtocolError:
		return []byte("(protocol error)")
	case WebSocketCloseUnsupportedData:
		return []byte("(unsupported data)")
	case WebSocketCloseNoStatusReceived:
		return []byte("(no status)")
	case WebSocketCloseAbnormal:
		return []byte("(abnormal closure)")
	case WebSocketCloseInvalidPayloadData:
		return []byte("(invalid payload data)")
	case WebSocketClosePolicyViolation:
		return []byte("(policy violation)")
	case WebSocketCloseMessageTooBig:
		return []byte("(message too big)")
	case WebSocketCloseMandatoryExtension:
		return []byte("(mandatory extension missing)")
	case WebSocketCloseInternalServerError:
		return []byte("(internal server error)")
	case WebSocketCloseServiceRestart:
		return []byte("(service restart)")
	case WebSocketCloseTryAgainLater:
		return []byte("(try again later)")
	case WebSocketCloseTLSHandshake:
		return []byte("(tls handshake error)")
	}
	return nil
}
