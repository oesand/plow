package specs

import (
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/internal/plain"
	"golang.org/x/net/publicsuffix"
	"iter"
	"sync"
	"time"
)

// NewCookieJar creates a new CookieJar instance.
func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make(map[string]map[string]*Cookie),
	}
}

// CookieJar is a thread-safe cookie storage that allows storing and retrieving cookies
// based on the host they are associated with. It uses the Effective TLD Plus One (eTLD+1) rule
// to determine the host for which the cookies are valid.
type CookieJar struct {
	cookies map[string]map[string]*Cookie

	mu sync.RWMutex
}

// GetCookie retrieves a cookie by its host and name.
//
// If the cookie is expired, it will be removed from the jar.
// If the host is not valid or the cookie does not exist, it returns nil.
func (jar *CookieJar) GetCookie(host string, name string) *Cookie {
	jar.mu.RLock()
	defer jar.mu.RUnlock()
	if jar.cookies == nil || len(jar.cookies) == 0 {
		return nil
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return nil
	}

	sub, has := jar.cookies[host]
	if !has {
		return nil
	}

	name = plain.TitleCase(name)
	value, has := sub[name]
	if has {
		if value.IsExpired(time.Now()) {
			delete(sub, name)
		} else {
			copied := *value
			return &copied
		}
	}
	return nil
}

// Cookies returns an iterator over all cookies associated with the given host.
func (jar *CookieJar) Cookies(host string) iter.Seq[Cookie] {
	jar.mu.RLock()
	defer jar.mu.RUnlock()
	if jar.cookies == nil || len(jar.cookies) == 0 {
		return internal.EmptyIterSeq[Cookie]()
	}

	tdl, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return internal.EmptyIterSeq[Cookie]()
	}

	sub, has := jar.cookies[tdl]
	if !has {
		return internal.EmptyIterSeq[Cookie]()
	}

	return func(yield func(Cookie) bool) {
		jar.mu.RLock()
		defer jar.mu.RUnlock()

		now := time.Now()
		var expired []string
		for subKey, cookie := range internal.IterMapSorted(sub) {
			if cookie.IsExpired(now) {
				expired = append(expired, subKey)
				continue
			}
			if !yield(*cookie) {
				break
			}
		}
		for _, key := range expired {
			delete(sub, key)
		}
	}
}

// SetCookie sets a single cookie for the specified host.
func (jar *CookieJar) SetCookie(host string, cookie Cookie) {
	jar.SetCookiesIter(host, func(yield func(Cookie) bool) {
		yield(cookie)
	})
}

// SetCookies sets multiple cookies for the specified host.
func (jar *CookieJar) SetCookies(host string, cookies []Cookie) {
	jar.SetCookiesIter(host, func(yield func(Cookie) bool) {
		for _, cookie := range cookies {
			if !yield(cookie) {
				break
			}
		}
	})
}

// SetCookiesIter sets multiple cookies for the specified host using an iterator.
func (jar *CookieJar) SetCookiesIter(host string, cookies iter.Seq[Cookie]) {
	jar.mu.Lock()
	defer jar.mu.Unlock()

	tdl, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return
	}

	if jar.cookies == nil {
		jar.cookies = make(map[string]map[string]*Cookie)
	}

	sub, has := jar.cookies[tdl]
	if !has {
		sub = map[string]*Cookie{}
		jar.cookies[tdl] = sub
	}

	now := time.Now()
	for cookie := range cookies {
		if cookie.IsExpired(now) {
			continue
		}

		sub[plain.TitleCase(cookie.Name)] = &cookie
	}
}
