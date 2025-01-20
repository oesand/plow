package specs

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

var (
	cookieDelimiter   = []byte("; ")
	cookieKeyExpires  = []byte("Expires")
	cookieKeyDomain   = []byte("Domain")
	cookieKeyPath     = []byte("Path")
	cookieKeyHTTPOnly = []byte("HttpOnly")
	cookieKeySecure   = []byte("Secure")
	cookieKeyMaxAge   = []byte("Max-Age")
	cookieKeySameSite = []byte("SameSite")

	directColonSpace = []byte(": ")
	directCrlf       = []byte("\r\n")
	headerCookie     = []byte("Cookie: ")
	headerSetCookie  = []byte("Set-Cookie: ")

	websocketAcceptBaseKey = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
)

type GigletOp string
type GigletError struct {
	Op  GigletOp
	Err error
}

func (e *GigletError) String() string {
	if e.Op != "" {
		return fmt.Sprintf("giglet/%s: %s", e.Op, e.Err)
	}
	return fmt.Sprintf("giglet: %s", e.Err)
}

func (e *GigletError) Error() string {
	return e.String()
}

func MatchError(err, other error) bool {
	gerr, _ := err.(*GigletError)
	ogerr, _ := err.(*GigletError)
	if gerr != nil || ogerr != nil {
		return gerr != nil && ogerr != nil &&
			gerr.Op == ogerr.Op && errors.Is(gerr.Err, ogerr.Err)
	}
	return errors.Is(err, other)
}

func ComputeWebSocketAcceptKey(challengeKey string) string {
	h := sha1.New() // (CWE-326) -- https://datatracker.ietf.org/doc/html/rfc6455#page-54
	h.Write([]byte(challengeKey))
	h.Write(websocketAcceptBaseKey)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func BasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString(internal.StringToBuffer(auth))
}
