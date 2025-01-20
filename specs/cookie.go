package specs

import (
	"bytes"
	"github.com/oesand/giglet/internal"
	"strconv"
	"time"
)

type CookieSameSite string

const (
	CookieSameSiteLaxMode    CookieSameSite = "Lax"
	CookieSameSiteStrictMode CookieSameSite = "Strict"
	CookieSameSiteNoneMode   CookieSameSite = "None"
)

type Cookie struct {
	Name  string
	Value string

	Domain  string
	MaxAge  uint64
	Expires time.Time
	Path    string

	HttpOnly bool
	Secure   bool
	SameSite CookieSameSite
}

func (cookie *Cookie) IsExpired(now time.Time) bool {
	if cookie.MaxAge > 0 {
		cookie.Expires = now.Add(time.Duration(cookie.MaxAge) * time.Second)
		cookie.MaxAge = 0
		return false
	}
	if cookie.MaxAge < 0 {
		return true
	}

	return !cookie.Expires.IsZero() && cookie.Expires.Before(now)
}

func (cookie *Cookie) Bytes(short bool) []byte {
	if short {
		var buf []byte

		buf = append(buf, internal.StringToBuffer(cookie.Name)...)
		buf = append(buf, '=')
		buf = append(buf, internal.StringToBuffer(cookie.Value)...)

		return buf
	}

	var buf bytes.Buffer

	if cookie.MaxAge > 0 {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeyMaxAge)
		buf.WriteByte('=')
		buf.Write(strconv.AppendUint(nil, cookie.MaxAge, 10))
	} else if !cookie.Expires.IsZero() {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeyExpires)
		buf.WriteByte('=')
		buf.Write(cookie.Expires.UTC().AppendFormat(nil, TimeFormat))
	}

	if len(cookie.Domain) > 0 {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeyDomain)
		buf.WriteByte('=')
		buf.WriteString(cookie.Domain)
	}

	if len(cookie.Path) > 0 {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeyPath)
		buf.WriteByte('=')
		buf.WriteString(cookie.Path)
	}

	if cookie.HttpOnly {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeyHTTPOnly)
	}

	if cookie.Secure {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeySecure)
	}

	if len(cookie.SameSite) > 0 {
		buf.Write(cookieDelimiter)
		buf.Write(cookieKeySameSite)
		buf.WriteByte('=')
		buf.WriteString(string(cookie.SameSite))
	}

	return buf.Bytes()
}
