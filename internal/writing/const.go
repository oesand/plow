package writing

import "github.com/oesand/giglet/specs"

var (
	writeHeadOp = specs.GigletOp("write/head")

	rawColonSpace      = []byte(": ")
	rawCookieDelimiter = []byte("; ")
	rawCookie          = []byte("Cookie: ")
	rawSetCookie       = []byte("Set-Cookie: ")
	rawCrlf            = []byte("\r\n")

	rawCookieKeyExpires  = []byte("Expires")
	rawCookieKeyDomain   = []byte("Domain")
	rawCookieKeyPath     = []byte("Path")
	rawCookieKeyHTTPOnly = []byte("HttpOnly")
	rawCookieKeySecure   = []byte("Secure")
	rawCookieKeyMaxAge   = []byte("Max-Age")
	rawCookieKeySameSite = []byte("SameSite")

	httpV10 = []byte("HTTP/1.0")
	httpV11 = []byte("HTTP/1.1")
)
