package specs

import (
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/plain"
	"strings"
)

// Query represents a parsed query string from a URL.
type Query map[string]string

// ParseQuery parses a query string into a Query map.
func ParseQuery(query string) Query {
	q := make(Query)
	if query == "" {
		return nil
	}

	pairs := strings.Split(query, "&")
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || key == "" {
			continue
		}

		decodedKey, err := plain.UnEscapeUrl(key, plain.EscapingQueryComponent)
		if err != nil {
			continue
		}
		if _, has := q[decodedKey]; has {
			continue
		}

		decodedValue, err := plain.UnEscapeUrl(value, plain.EscapingQueryComponent)
		if err != nil {
			continue
		}

		q[decodedKey] = decodedValue
	}

	return q
}

// Any checks if the Query contains any key-value pairs.
func (q Query) Any() bool {
	return q != nil && len(q) > 0
}

// String returns the query string representation of the Query.
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
