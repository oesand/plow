package specs

import (
	"fmt"
	"net/url"
	"strings"
)

type Form url.Values
type Query map[string]string

func ParseQuery(query string) (q Query, err error) {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = fmt.Errorf("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}

		key, value, _ := strings.Cut(key, "=")
		key, err = url.QueryUnescape(key)
		if _, has := q[key]; err != nil || has {
			continue
		}
		value, err = url.QueryUnescape(value)
		if err != nil {
			continue
		}
		q[key] = value
	}
	return q, err
}

func (q Query) String() string {
	if len(q) == 0 {
		return ""
	}
	var buf strings.Builder
	for k, v := range q {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
	}
	return buf.String()
}
