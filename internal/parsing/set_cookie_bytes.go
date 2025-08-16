package parsing

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"strconv"
)

var (
	rawCookieDelimiter = []byte("; ")

	rawCookieKeyExpires  = []byte("Expires")
	rawCookieKeyDomain   = []byte("Domain")
	rawCookieKeyPath     = []byte("Path")
	rawCookieKeyHTTPOnly = []byte("HttpOnly")
	rawCookieKeySecure   = []byte("Secure")
	rawCookieKeyMaxAge   = []byte("Max-Age")
	rawCookieKeySameSite = []byte("SameSite")
)

func SetCookieBytes(cookie *specs.Cookie) []byte {
	var buf bytes.Buffer

	buf.WriteString(cookie.Name)
	buf.WriteByte('=')
	buf.WriteString(cookie.Value)

	if cookie.MaxAge > 0 {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeyMaxAge)
		buf.WriteByte('=')
		buf.Write(strconv.AppendUint(nil, cookie.MaxAge, 10))
	} else if !cookie.Expires.IsZero() {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeyExpires)
		buf.WriteByte('=')
		buf.Write(cookie.Expires.UTC().AppendFormat(nil, specs.TimeFormat))
	}

	if len(cookie.Domain) > 0 {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeyDomain)
		buf.WriteByte('=')
		buf.WriteString(cookie.Domain)
	}

	if len(cookie.Path) > 0 {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeyPath)
		buf.WriteByte('=')
		buf.WriteString(cookie.Path)
	}

	if cookie.HttpOnly {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeyHTTPOnly)
	}

	if cookie.Secure {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeySecure)
	}

	if len(cookie.SameSite) > 0 {
		buf.Write(rawCookieDelimiter)
		buf.Write(rawCookieKeySameSite)
		buf.WriteByte('=')
		buf.WriteString(string(cookie.SameSite))
	}

	return buf.Bytes()
}
