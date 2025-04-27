package writing

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"io"
)

func WriteRequestHead(writer io.Writer, method specs.HttpMethod, url *specs.Url, header *specs.Header) (int64, error) {
	// Headline
	buf := bytes.NewBufferString(string(method))
	buf.WriteRune(' ')
	buf.WriteString(url.Path)

	query := url.Query()
	if query != "" {
		buf.WriteRune('?')
		buf.WriteString(query)
	}

	buf.WriteRune(' ')
	buf.Write(httpV11)

	buf.Write(rawCrlf)

	// Headers
	for key, value := range header.All() {
		buf.WriteString(key)
		buf.Write(rawColonSpace)
		buf.WriteString(value)
		buf.Write(rawCrlf)
	}

	if header.HasCookies() {
		buf.Write(rawCookie)

		firstCookie := true
		for cookie := range header.Cookies() {
			if firstCookie {
				firstCookie = false
			} else {
				buf.Write(rawCookieDelimiter)
			}
			buf.WriteString(cookie.Name)
			buf.WriteRune('=')
			buf.WriteString(cookie.Value)
		}

		buf.Write(rawCrlf)
	}

	buf.Write(rawCrlf)

	i, err := buf.WriteTo(writer)
	if err != nil {
		return -1, &specs.GigletError{
			Op:  writeHeadOp,
			Err: err,
		}
	}
	return i, nil
}
