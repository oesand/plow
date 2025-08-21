package specs

import (
	"iter"
	"maps"

	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/internal/plain"
)

// NewHeader creates a new Header instance with optional configuration functions.
func NewHeader(configure ...func(header *Header)) *Header {
	header := &Header{}

	for _, conf := range configure {
		conf(header)
	}

	return header
}

// Header represents a collection of HTTP headers and cookies.
// It provides methods to set, get, and manipulate headers and cookies.
type Header struct {
	headers map[string]string
	cookies map[string]*Cookie
}

// Clone creates a deep copy of the Header instance.
func (header *Header) Clone() *Header {
	return &Header{
		headers: maps.Clone(header.headers),
		cookies: maps.Clone(header.cookies),
	}
}

// Any checks if the Header contains any headers.
func (header *Header) Any() bool {
	return header.headers != nil && len(header.headers) > 0
}

// Get retrieves the value of a header by its name.
func (header *Header) Get(name string) string {
	value, _ := header.TryGet(name)
	return value
}

// TryGet attempts to retrieve the value of a header by its name.
func (header *Header) TryGet(name string) (string, bool) {
	if header.Any() {
		value, has := header.headers[plain.TitleCase(name)]
		return value, has
	}
	return "", false
}

// Has checks if a header with the specified name exists.
func (header *Header) Has(name string) bool {
	if header.Any() {
		_, has := header.headers[plain.TitleCase(name)]
		return has
	}
	return false
}

// Set adds or updates a header with the specified name and value.
func (header *Header) Set(name, value string) {
	name = plain.TitleCase(name)
	if name == "Set-Cookie" || name == "Cookie" {
		panic("plow: header not support direct set cookie, use method 'SetCookie'")
	} else if header.headers == nil {
		header.headers = map[string]string{}
	}
	header.headers[name] = value
}

// Del removes a header by its name.
func (header *Header) Del(name string) {
	if header.Any() {
		delete(header.headers, plain.TitleCase(name))
	}
}

// All returns an iterator over all headers in the Header instance.
func (header *Header) All() iter.Seq2[string, string] {
	if !header.Any() {
		return internal.EmptyIterSeq2[string, string]()
	}
	return internal.IterMapSorted(header.headers)
}

// AnyCookies checks if the Header contains any cookies.
func (header *Header) AnyCookies() bool {
	return header.cookies != nil && len(header.cookies) > 0
}

// GetCookie retrieves a cookie by its name.
func (header *Header) GetCookie(name string) *Cookie {
	if header.AnyCookies() {
		val, _ := header.cookies[plain.TitleCase(name)]
		return val
	}
	return nil
}

// HasCookie checks if a cookie with the specified name exists.
func (header *Header) HasCookie(name string) bool {
	if header.AnyCookies() {
		_, has := header.cookies[plain.TitleCase(name)]
		return has
	}
	return false
}

// DelCookie removes a cookie by its name.
func (header *Header) DelCookie(name string) {
	if header.AnyCookies() {
		delete(header.cookies, plain.TitleCase(name))
	}
}

// SetCookie adds or updates a cookie in the Header instance.
func (header *Header) SetCookie(cookie Cookie) {
	if cookie.Name == "" {
		return
	}
	if header.cookies == nil {
		header.cookies = map[string]*Cookie{}
	}

	header.cookies[plain.TitleCase(cookie.Name)] = &cookie
}

// SetCookieValue sets a cookie with the specified name and value.
func (header *Header) SetCookieValue(name, value string) {
	header.SetCookie(Cookie{
		Name:  name,
		Value: value,
	})
}

// Cookies returns an iterator over all cookies in the Header instance.
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
