package specs

import (
	"bytes"
	"giglet/internal"
	"iter"
)

type Header struct {
	_ internal.NoCopy

	headers map[string]string
	cookies map[string]*Cookie
}

func (header *Header) Get(name string) string {
	if header.headers == nil {
		return ""
	}
	return header.headers[name]
}

func (header *Header) Has(name string) bool {
	if header.headers == nil {
		return false
	}
	_, has := header.headers[name]
	return has
}

func (header *Header) Set(name, value string) {
	name = internal.TitleCase(name)
	if name == "Set-Cookie" {
		panic("header not support direct set cookie, use method 'SetCookie'")
	} else if header.headers == nil {
		header.headers = map[string]string{}
	}
	header.headers[name] = value
}

func (header *Header) Del(name string) {
	if header.headers != nil {
		delete(header.headers, internal.TitleCase(name))
	}
}

func (header *Header) All() iter.Seq2[string, string] {
	if header.headers == nil {
		return internal.EmptyIterSeq2[string, string]()
	}

	return func(yield func(string, string) bool) {
		for name, value := range header.headers {
			if !yield(name, value) {
				break
			}
		}
	}
}

func (header *Header) GetCookie(name string) *Cookie {
	if header.cookies == nil {
		return nil
	}
	return header.cookies[name]
}

func (header *Header) HasCookie(name string) bool {
	if header.cookies == nil {
		return false
	}
	_, has := header.cookies[name]
	return has
}

func (header *Header) DelCookie(name string) {
	if header.cookies != nil {
		delete(header.cookies, internal.TitleCase(name))
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
	if header.cookies == nil {
		return internal.EmptyIterSeq[Cookie]()
	}

	return func(yield func(Cookie) bool) {
		for _, cookie := range header.cookies {
			if !yield(*cookie) {
				break
			}
		}
	}
}

func (header *Header) Bytes() []byte {
	if header.headers == nil || len(header.headers) == 0 {
		return make([]byte, 0)
	}
	var buf bytes.Buffer

	for key, value := range header.headers {
		buf.Write(internal.StringToBuffer(key))
		buf.Write(directColonSpace)
		buf.Write(internal.StringToBuffer(value))
		buf.Write(directCrlf)
	}

	return buf.Bytes()
}

func (header *Header) SetCookieHeaderBytes() []byte {
	if header.cookies == nil || len(header.cookies) == 0 {
		return make([]byte, 0)
	}

	var buf bytes.Buffer

	for _, cookie := range header.cookies {
		buf.Write(headerSetCookie)
		buf.Write(cookie.Bytes(false))
		buf.Write(directCrlf)
	}

	return buf.Bytes()
}

func (header *Header) CookieHeaderBytes() []byte {
	if header.cookies == nil || len(header.cookies) == 0 {
		return make([]byte, 0)
	}

	var buf bytes.Buffer

	buf.Write(headerCookie)

	first := true
	for _, cookie := range header.cookies {
		if first {
			first = false
		} else {
			buf.Write(cookieDelimiter)
		}
		buf.Write(cookie.Bytes(true))
	}
	buf.Write(directCrlf)

	return buf.Bytes()
}
