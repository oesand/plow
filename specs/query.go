package specs

import (
	"github.com/oesand/giglet/internal/utils"
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
	if len(q) == 0 {
		return ""
	}
	var buf strings.Builder
	for k, v := range utils.IterMapSorted(q) {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(utils.EscapeUrl(k, utils.EscapingQueryComponent))
		buf.WriteByte('=')
		buf.WriteString(utils.EscapeUrl(v, utils.EscapingQueryComponent))
	}
	return buf.String()
}
