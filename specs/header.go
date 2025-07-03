package specs

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/plain"
	"iter"
)

func NewHeader(configure ...func(header *Header)) *Header {
	header := &Header{}

	for _, conf := range configure {
		conf(header)
	}

	return header
}

type Header struct {
	_ internal.NoCopy

	headers map[string]string
	cookies map[string]*Cookie
}

func (header *Header) Any() bool {
	return header.headers != nil && len(header.headers) > 0
}

func (header *Header) Get(name string) string {
	value, _ := header.TryGet(name)
	return value
}

func (header *Header) TryGet(name string) (string, bool) {
	if header.Any() {
		value, has := header.headers[plain.TitleCase(name)]
		return value, has
	}
	return "", false
}

func (header *Header) Has(name string) bool {
	if header.Any() {
		_, has := header.headers[plain.TitleCase(name)]
		return has
	}
	return false
}

func (header *Header) Set(name, value string) {
	name = plain.TitleCase(name)
	if name == "Set-Cookie" || name == "Cookie" {
		panic("header not support direct set cookie, use method 'SetCookie'")
	} else if header.headers == nil {
		header.headers = map[string]string{}
	}
	header.headers[name] = value
}

func (header *Header) Del(name string) {
	if header.Any() {
		delete(header.headers, plain.TitleCase(name))
	}
}

func (header *Header) All() iter.Seq2[string, string] {
	if !header.Any() {
		return internal.EmptyIterSeq2[string, string]()
	}
	return internal.IterMapSorted(header.headers)
}

func (header *Header) AnyCookies() bool {
	return header.cookies != nil && len(header.cookies) > 0
}

func (header *Header) GetCookie(name string) *Cookie {
	if header.AnyCookies() {
		return header.cookies[name]
	}
	return nil
}

func (header *Header) HasCookie(name string) bool {
	if header.AnyCookies() {
		_, has := header.cookies[name]
		return has
	}
	return false
}

func (header *Header) DelCookie(name string) {
	if header.AnyCookies() {
		delete(header.cookies, plain.TitleCase(name))
	}
}

func (header *Header) SetCookie(cookie Cookie) {
	if cookie.Name == "" {
		return
	}
	if header.cookies == nil {
		header.cookies = map[string]*Cookie{}
	}

	header.cookies[cookie.Name] = &cookie
}

func (header *Header) SetCookieValue(name, value string) {
	header.SetCookie(Cookie{
		Name:  name,
		Value: value,
	})
}

func (header *Header) Cookies() iter.Seq[Cookie] {
	if !header.AnyCookies() {
		return internal.EmptyIterSeq[Cookie]()
	}

	if internal.IsNotTesting {
		return func(yield func(Cookie) bool) {
			for _, cookie := range header.cookies {
				if !yield(*cookie) {
					break
				}
			}
		}
	}

	keys := internal.IterKeysSorted(header.cookies)
	return func(yield func(Cookie) bool) {
		for k := range keys {
			if !yield(*header.cookies[k]) {
				break
			}
		}
	}
}
