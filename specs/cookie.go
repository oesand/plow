package specs

import (
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

	return !cookie.Expires.IsZero() && now.After(cookie.Expires)
}
