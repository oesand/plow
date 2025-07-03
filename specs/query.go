package specs

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/plain"
	"net/url"
	"strings"
)

type Query map[string]string

func ParseQuery(query string) Query {
	q := make(Query)
	if query == "" {
		return q
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			continue
		}

		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			continue
		}
		if _, has := q[decodedKey]; has {
			continue
		}

		decodedValue, err := url.QueryUnescape(value)
		if err != nil {
			continue
		}

		q[decodedKey] = decodedValue
	}

	return q
}

func (q Query) String() string {
	if q == nil || len(q) == 0 {
		return ""
	}
	var buf strings.Builder
	for k, v := range internal.IterMapSorted(q) {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(plain.EscapeUrl(k, plain.EscapingQueryComponent))
		buf.WriteByte('=')
		buf.WriteString(plain.EscapeUrl(v, plain.EscapingQueryComponent))
	}
	return buf.String()
}
