package specs

import (
	"fmt"
	"github.com/oesand/giglet/internal/utils"
	"golang.org/x/net/publicsuffix"
	"iter"
	"sync"
	"time"
)

func cookiejarSubkey(cookie *Cookie) string {
	return fmt.Sprintf("%s|%s|%s", cookie.Name, cookie.Domain, cookie.Path)
}

type CookieJar struct {
	mutex sync.RWMutex

	cookies map[string]map[string]*Cookie
}

func (jar *CookieJar) GetCookie(url *Url, name string) *Cookie {
	host, err := publicsuffix.EffectiveTLDPlusOne(url.Host)
	if err != nil {
		return nil
	}

	sub, has := jar.cookies[host]
	if !has {
		return nil
	}

	key := fmt.Sprintf("%s|%s|%s", name, "."+host, "/")
	value, has := sub[key]
	if !has {
		key = fmt.Sprintf("%s|%s|%s", name, host, "/")
		value, has = sub[key]
	}

	if has && value.IsExpired(time.Now()) {
		delete(sub, key)
		return nil
	}
	return value
}

func (jar *CookieJar) Cookies(url *Url) iter.Seq[Cookie] {
	jar.mutex.RLock()
	defer jar.mutex.RUnlock()

	host, err := publicsuffix.EffectiveTLDPlusOne(url.Host)
	if err != nil {
		return utils.EmptyIterSeq[Cookie]()
	}

	sub, has := jar.cookies[host]
	if !has {
		return utils.EmptyIterSeq[Cookie]()
	}

	return func(yield func(Cookie) bool) {
		jar.mutex.RLock()
		defer jar.mutex.RUnlock()

		now := time.Now()
		expired := []string{}
		for subkey, cookie := range utils.IterMapSorted(sub) {
			if cookie.IsExpired(now) {
				expired = append(expired, subkey)
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

func (jar *CookieJar) SetCookie(url *Url, cookie Cookie) {
	jar.mutex.Lock()
	defer jar.mutex.Unlock()

	now := time.Now()
	if cookie.IsExpired(now) {
		return
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(url.Host)
	if err != nil {
		return
	}

	if jar.cookies == nil {
		jar.cookies = map[string]map[string]*Cookie{}
	}

	sub, has := jar.cookies[host]
	if !has {
		sub = map[string]*Cookie{}
		jar.cookies[host] = sub
	}
	if cookie.Path == "" {
		cookie.Path = "/"
	}

	sub[cookiejarSubkey(&cookie)] = &cookie
}

func (jar *CookieJar) SetCookies(url *Url, cookies []Cookie) {
	jar.SetCookiesIter(url, func(yield func(Cookie) bool) {
		for _, cookie := range cookies {
			if !yield(cookie) {
				break
			}
		}
	})
}

func (jar *CookieJar) SetCookiesIter(url *Url, cookies iter.Seq[Cookie]) {
	jar.mutex.Lock()
	defer jar.mutex.Unlock()

	key, err := publicsuffix.EffectiveTLDPlusOne(url.Host)
	if err != nil {
		return
	}

	if jar.cookies == nil {
		jar.cookies = make(map[string]map[string]*Cookie)
	}

	sub, has := jar.cookies[key]
	if !has {
		sub = map[string]*Cookie{}
		jar.cookies[key] = sub
	}

	now := time.Now()
	for cookie := range cookies {
		if cookie.IsExpired(now) {
			continue
		}
		if cookie.Path == "" {
			cookie.Path = "/"
		}

		sub[cookiejarSubkey(&cookie)] = &cookie
	}
}
