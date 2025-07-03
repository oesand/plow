package specs

import (
	"github.com/oesand/giglet/internal"
	"golang.org/x/net/publicsuffix"
	"iter"
	"sync"
	"time"
)

type CookieJar struct {
	mutex sync.RWMutex

	cookies map[string]map[string]*Cookie
}

func (jar *CookieJar) GetCookie(host string, name string) *Cookie {
	jar.mutex.RLock()
	defer jar.mutex.RUnlock()
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

func (jar *CookieJar) Cookies(host string) iter.Seq[Cookie] {
	jar.mutex.RLock()
	defer jar.mutex.RUnlock()
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
		jar.mutex.RLock()
		defer jar.mutex.RUnlock()

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

func (jar *CookieJar) SetCookie(host string, cookie Cookie) {
	jar.SetCookiesIter(host, func(yield func(Cookie) bool) {
		yield(cookie)
	})
}

func (jar *CookieJar) SetCookies(host string, cookies []Cookie) {
	jar.SetCookiesIter(host, func(yield func(Cookie) bool) {
		for _, cookie := range cookies {
			if !yield(cookie) {
				break
			}
		}
	})
}

func (jar *CookieJar) SetCookiesIter(host string, cookies iter.Seq[Cookie]) {
	jar.mutex.Lock()
	defer jar.mutex.Unlock()

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

		sub[cookie.Name] = &cookie
	}
}
