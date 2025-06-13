package specs

import "strconv"

type StatusCode uint16

const (
	StatusCodeUndefined StatusCode = 0

	StatusCodeContinue           StatusCode = 100
	StatusCodeSwitchingProtocols StatusCode = 101
	StatusCodeProcessing         StatusCode = 102
	StatusCodeEarlyHints         StatusCode = 103

	StatusCodeOK                   StatusCode = 200
	StatusCodeCreated              StatusCode = 201
	StatusCodeAccepted             StatusCode = 202
	StatusCodeNonAuthoritativeInfo StatusCode = 203
	StatusCodeNoContent            StatusCode = 204
	StatusCodeResetContent         StatusCode = 205
	StatusCodePartialContent       StatusCode = 206
	StatusCodeMultiStatus          StatusCode = 207
	StatusCodeAlreadyReported      StatusCode = 208
	StatusCodeIMUsed               StatusCode = 226

	StatusCodeMultipleChoices   StatusCode = 300
	StatusCodeMovedPermanently  StatusCode = 301
	StatusCodeFound             StatusCode = 302
	StatusCodeSeeOther          StatusCode = 303
	StatusCodeNotModified       StatusCode = 304
	StatusCodeUseProxy          StatusCode = 305
	StatusCodeTemporaryRedirect StatusCode = 307
	StatusCodePermanentRedirect StatusCode = 308

	StatusCodeBadRequest                   StatusCode = 400
	StatusCodeUnauthorized                 StatusCode = 401
	StatusCodePaymentRequired              StatusCode = 402
	StatusCodeForbidden                    StatusCode = 403
	StatusCodeNotFound                     StatusCode = 404
	StatusCodeMethodNotAllowed             StatusCode = 405
	StatusCodeNotAcceptable                StatusCode = 406
	StatusCodeProxyAuthRequired            StatusCode = 407
	StatusCodeRequestTimeout               StatusCode = 408
	StatusCodeConflict                     StatusCode = 409
	StatusCodeGone                         StatusCode = 410
	StatusCodeLengthRequired               StatusCode = 411
	StatusCodePreconditionFailed           StatusCode = 412
	StatusCodeRequestEntityTooLarge        StatusCode = 413
	StatusCodeRequestURITooLong            StatusCode = 414
	StatusCodeUnsupportedMediaType         StatusCode = 415
	StatusCodeRequestedRangeNotSatisfiable StatusCode = 416
	StatusCodeExpectationFailed            StatusCode = 417
	StatusCodeTeapot                       StatusCode = 418
	StatusCodeMisdirectedRequest           StatusCode = 421
	StatusCodeUnprocessableEntity          StatusCode = 422
	StatusCodeLocked                       StatusCode = 423
	StatusCodeFailedDependency             StatusCode = 424
	StatusCodeTooEarly                     StatusCode = 425
	StatusCodeUpgradeRequired              StatusCode = 426
	StatusCodePreconditionRequired         StatusCode = 428
	StatusCodeTooManyRequests              StatusCode = 429
	StatusCodeRequestHeaderFieldsTooLarge  StatusCode = 431
	StatusCodeUnavailableForLegalReasons   StatusCode = 451

	StatusCodeInternalServerError           StatusCode = 500
	StatusCodeNotImplemented                StatusCode = 501
	StatusCodeBadGateway                    StatusCode = 502
	StatusCodeServiceUnavailable            StatusCode = 503
	StatusCodeGatewayTimeout                StatusCode = 504
	StatusCodeHTTPVersionNotSupported       StatusCode = 505
	StatusCodeVariantAlsoNegotiates         StatusCode = 506
	StatusCodeInsufficientStorage           StatusCode = 507
	StatusCodeLoopDetected                  StatusCode = 508
	StatusCodeNotExtended                   StatusCode = 510
	StatusCodeNetworkAuthenticationRequired StatusCode = 511
)

func (status StatusCode) IsReplyable() bool {
	noContent := (100 <= status && status < 200) || status == 204 || (300 <= status && status < 400)
	return !noContent
}

func (status StatusCode) IsValid() bool {
	return 100 <= status && status < 600
}

func (status StatusCode) IsRedirect() bool {
	return status == StatusCodeMovedPermanently ||
		status == StatusCodeFound ||
		status == StatusCodeSeeOther ||
		status == StatusCodeTemporaryRedirect ||
		status == StatusCodePermanentRedirect
}

func (status StatusCode) Formatted() []byte {
	buf := strconv.AppendUint(nil, uint64(status), 10)
	buf = append(buf, ' ')
	return append(buf, status.Detail()...)
}

func (code StatusCode) Detail() []byte {
	switch code {
	case StatusCodeContinue:
		return []byte("Continue")
	case StatusCodeSwitchingProtocols:
		return []byte("Switching Protocols")
	case StatusCodeProcessing:
		return []byte("Processing")
	case StatusCodeEarlyHints:
		return []byte("Early Hints")
	case StatusCodeOK:
		return []byte("OK")
	case StatusCodeCreated:
		return []byte("Created")
	case StatusCodeAccepted:
		return []byte("Accepted")
	case StatusCodeNonAuthoritativeInfo:
		return []byte("Non-Authoritative Information")
	case StatusCodeNoContent:
		return []byte("No Content")
	case StatusCodeResetContent:
		return []byte("Reset Content")
	case StatusCodePartialContent:
		return []byte("Partial Content")
	case StatusCodeMultiStatus:
		return []byte("Multi-Status")
	case StatusCodeAlreadyReported:
		return []byte("Already Reported")
	case StatusCodeIMUsed:
		return []byte("I'm Used")
	case StatusCodeMultipleChoices:
		return []byte("Multiple Choices")
	case StatusCodeMovedPermanently:
		return []byte("Moved Permanently")
	case StatusCodeFound:
		return []byte("Found")
	case StatusCodeSeeOther:
		return []byte("See Other")
	case StatusCodeNotModified:
		return []byte("Not Modified")
	case StatusCodeUseProxy:
		return []byte("Use Proxy")
	case StatusCodeTemporaryRedirect:
		return []byte("Temporary Redirect")
	case StatusCodePermanentRedirect:
		return []byte("Permanent Redirect")
	case StatusCodeBadRequest:
		return []byte("Bad Request")
	case StatusCodeUnauthorized:
		return []byte("Unauthorized")
	case StatusCodePaymentRequired:
		return []byte("Payment Required")
	case StatusCodeForbidden:
		return []byte("Forbidden")
	case StatusCodeNotFound:
		return []byte("Not Found")
	case StatusCodeMethodNotAllowed:
		return []byte("Method Not Allowed")
	case StatusCodeNotAcceptable:
		return []byte("Not Acceptable")
	case StatusCodeProxyAuthRequired:
		return []byte("Proxy Auth Required")
	case StatusCodeRequestTimeout:
		return []byte("Request Timeout")
	case StatusCodeConflict:
		return []byte("Conflict")
	case StatusCodeGone:
		return []byte("Gone")
	case StatusCodeLengthRequired:
		return []byte("Length Required")
	case StatusCodePreconditionFailed:
		return []byte("Precondition Failed")
	case StatusCodeRequestEntityTooLarge:
		return []byte("Request Entity Too Large")
	case StatusCodeRequestURITooLong:
		return []byte("Request URI Too Long")
	case StatusCodeUnsupportedMediaType:
		return []byte("Unsupported Media Type")
	case StatusCodeRequestedRangeNotSatisfiable:
		return []byte("Requested Range Not Satisfiable")
	case StatusCodeExpectationFailed:
		return []byte("ExpectationFailed")
	case StatusCodeTeapot:
		return []byte("I'm a teapot")
	case StatusCodeMisdirectedRequest:
		return []byte("Misdirected Request")
	case StatusCodeUnprocessableEntity:
		return []byte("Unprocessable Entity")
	case StatusCodeLocked:
		return []byte("Locked")
	case StatusCodeFailedDependency:
		return []byte("Failed Dependency")
	case StatusCodeTooEarly:
		return []byte("Too Early")
	case StatusCodeUpgradeRequired:
		return []byte("Upgrade Required")
	case StatusCodePreconditionRequired:
		return []byte("Precondition Required")
	case StatusCodeTooManyRequests:
		return []byte("Too Many Requests")
	case StatusCodeRequestHeaderFieldsTooLarge:
		return []byte("Request Header Fields Too Large")
	case StatusCodeUnavailableForLegalReasons:
		return []byte("Unavailable For Legal Reasons")
	case StatusCodeInternalServerError:
		return []byte("Internal Server Error")
	case StatusCodeNotImplemented:
		return []byte("Not Implemented")
	case StatusCodeBadGateway:
		return []byte("Bad Gateway")
	case StatusCodeServiceUnavailable:
		return []byte("Service Unavailable")
	case StatusCodeGatewayTimeout:
		return []byte("Gateway Timeout")
	case StatusCodeHTTPVersionNotSupported:
		return []byte("HTTP Version Not Supported")
	case StatusCodeVariantAlsoNegotiates:
		return []byte("Variant Also Negotiates")
	case StatusCodeInsufficientStorage:
		return []byte("Insufficient Storage")
	case StatusCodeLoopDetected:
		return []byte("Loop Detected")
	case StatusCodeNotExtended:
		return []byte("Not Extended")
	case StatusCodeNetworkAuthenticationRequired:
		return []byte("Network Authentication Required")
	}
	return make([]byte, 0)
}
