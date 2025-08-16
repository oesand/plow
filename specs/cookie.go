package specs

import (
	"time"
)

// CookieSameSite defines the SameSite attribute for cookies.
type CookieSameSite string

// CookieSameSite modes define how cookies are sent with cross-site requests.
const (
	CookieSameSiteLaxMode    CookieSameSite = "Lax"
	CookieSameSiteStrictMode CookieSameSite = "Strict"
	CookieSameSiteNoneMode   CookieSameSite = "None"
)

// Cookie represents an HTTP cookie with its attributes.
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

// IsExpired checks if the cookie is expired based on the current time.
// If MaxAge is set, it will update the Expires time accordingly.
// If MaxAge is negative, it indicates the cookie is already expired.
// If Expires is set and the current time is after Expires, the cookie is expired
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
