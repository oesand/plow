package specs

import (
	"reflect"
	"testing"
)

func TestParseUrl(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    *Url
		invalid bool
	}{
		// Invalid url
		{
			name:    "Only username mark",
			raw:     "@",
			invalid: true,
		},
		{
			name:    "Only scheme mark",
			raw:     "://",
			invalid: true,
		},
		{
			name:    "Scheme & path with empty host",
			raw:     "http:///path",
			invalid: true,
		},
		{
			name:    "Only scheme",
			raw:     "http://",
			invalid: true,
		},
		{
			name:    "Only port",
			raw:     ":80",
			invalid: true,
		},
		{
			name:    "Empty port",
			raw:     "host:",
			invalid: true,
		},
		{
			name:    "Only username",
			raw:     "username@",
			invalid: true,
		},
		{
			name:    "Empty username with host",
			raw:     "@host",
			invalid: true,
		},
		{
			name:    "Empty username with host & port",
			raw:     "@host:80",
			invalid: true,
		},
		{
			name:    "Username & password only",
			raw:     "username:password@",
			invalid: true,
		},
		{
			name:    "Empty username with password only",
			raw:     ":password@",
			invalid: true,
		},
		{
			name:    "Empty username with password & host",
			raw:     ":password@host",
			invalid: true,
		},
		{
			name:    "Empty username with password & host & port",
			raw:     ":password@host:80",
			invalid: true,
		},
		{
			name:    "IPv6 without leading bracket",
			raw:     "2001:db8::1]",
			invalid: true,
		},
		{
			name:    "IPv6 without ending bracket",
			raw:     "[2001:db8::1",
			invalid: true,
		},
		{
			name:    "IPv6 without port",
			raw:     "[2001:db8::1]:",
			invalid: true,
		},

		// Single parts
		{
			name: "Empty string",
			raw:  "",
			want: &Url{},
		},
		{
			name: "Only host",
			raw:  "host",
			want: &Url{Host: "host"},
		},
		{
			name: "Only path",
			raw:  "/path",
			want: &Url{Path: "/path"},
		},
		{
			name: "Only fragment",
			raw:  "#fragment",
			want: &Url{Fragment: "fragment"},
		},
		{
			name: "Only query",
			raw:  "?key=value",
			want: &Url{Query: Query{"key": "value"}},
		},

		// Combined parts
		{
			name: "Basic HTTP",
			raw:  "http://example.com",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
			},
		},
		{
			name: "Basic HTTP with slash path",
			raw:  "http://example.com/",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
		},
		{
			name: "Basic HTTP with query mark",
			raw:  "http://example.com?",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
			},
		},
		{
			name: "Basic HTTP with slash path and query mark",
			raw:  "http://example.com/?",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/",
			},
		},
		{
			name: "Host with slash path",
			raw:  "example.com/",
			want: &Url{
				Host: "example.com",
				Path: "/",
			},
		},
		{
			name: "Host with query mark",
			raw:  "example.com?",
			want: &Url{
				Host: "example.com",
			},
		},
		{
			name: "Host with slash path and query mark",
			raw:  "example.com/?",
			want: &Url{
				Host: "example.com",
				Path: "/",
			},
		},
		{
			name: "Host & port with slash path",
			raw:  "example.com:80/",
			want: &Url{
				Host: "example.com",
				Port: 80,
				Path: "/",
			},
		},
		{
			name: "Host & port with query mark",
			raw:  "example.com:80?",
			want: &Url{
				Host: "example.com",
				Port: 80,
			},
		},
		{
			name: "Host & port with slash path and query mark",
			raw:  "example.com:80/?",
			want: &Url{
				Host: "example.com",
				Port: 80,
				Path: "/",
			},
		},
		{
			name: "Path and query mark",
			raw:  "/path?",
			want: &Url{
				Path: "/path",
			},
		},
		{
			name: "Path slash and query mark",
			raw:  "/?",
			want: &Url{
				Path: "/",
			},
		},
		{
			name: "Path and query",
			raw:  "/user?id=120",
			want: &Url{
				Path: "/user",
				Query: Query{
					"id": "120",
				},
			},
		},
		{
			name: "HTTPS with port",
			raw:  "http://example.com:8090",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
				Port:   8090,
			},
		},
		{
			name: "Host with port",
			raw:  "example.com:8090",
			want: &Url{
				Host: "example.com",
				Port: 8090,
			},
		},
		{
			name:    "Invalid port",
			raw:     "example.com:2j0",
			invalid: true,
		},
		{
			name: "Scheme with special characters",
			raw:  "ftp+rt9://example.com",
			want: &Url{
				Scheme: "ftp+rt9",
				Host:   "example.com",
			},
		},
		{
			name: "Basic HTTP with username",
			raw:  "http://username@example.com",
			want: &Url{
				Scheme:   "http",
				Username: "username",
				Host:     "example.com",
			},
		},
		{
			name: "Basic HTTPS with username and password",
			raw:  "https://username:password@example.com",
			want: &Url{
				Scheme:   "https",
				Username: "username",
				Password: "password",
				Host:     "example.com",
			},
		},
		{
			name: "HTTPS with path",
			raw:  "https://example.com/path",
			want: &Url{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/path",
			},
		},
		{
			name: "HTTPS with slash path",
			raw:  "https://example.com/",
			want: &Url{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/",
			},
		},
		{
			name: "HTTPS with path and fragment",
			raw:  "https://example.com/path#section",
			want: &Url{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/path",
				Fragment: "section",
			},
		},
		{
			name: "HTTPS with port and path and fragment",
			raw:  "https://example.com:8090/path#section",
			want: &Url{
				Scheme:   "https",
				Host:     "example.com",
				Port:     8090,
				Path:     "/path",
				Fragment: "section",
			},
		},
		{
			name: "HTTPS with query",
			raw:  "https://example.com?q=test&lang=en",
			want: &Url{
				Scheme: "https",
				Host:   "example.com",
				Query: Query{
					"q":    "test",
					"lang": "en",
				},
			},
		},
		{
			name: "URL with port and query",
			raw:  "http://example.com:8080/search?q=test&lang=en",
			want: &Url{
				Scheme: "http",
				Host:   "example.com",
				Port:   8080,
				Path:   "/search",
				Query: Query{
					"q":    "test",
					"lang": "en",
				},
			},
		},
		{
			name: "Encoded query string",
			raw:  "https://example.com/search?q=hello%20world",
			want: &Url{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/search",
				Query:  Query{"q": "hello world"},
			},
		},
		{
			name: "IPv6 with path with special characters and query",
			raw:  "http://[2001:db8::1]/file/a%20b?query=value",
			want: &Url{
				Scheme: "http",
				Host:   "[2001:db8::1]",
				Path:   "/file/a b",
				Query:  Query{"query": "value"},
			},
		},
		{
			name: "IPv6 with port and path with special characters",
			raw:  "http://[2001:db8::1]:8080/file/%C3%BC",
			want: &Url{
				Scheme: "http",
				Host:   "[2001:db8::1]",
				Port:   8080,
				Path:   "/file/ü",
			},
		},
		{
			name: "Full URL with all fields",
			raw:  "https://user:pass@my.example.com:8443/api/v1?key=value#anchor",
			want: &Url{
				Scheme:   "https",
				Username: "user",
				Password: "pass",
				Host:     "my.example.com",
				Port:     8443,
				Path:     "/api/v1",
				Fragment: "anchor",
				Query:    Query{"key": "value"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUrl(tt.raw)
			if tt.invalid {
				if err == nil {
					t.Errorf("ParseUrl() expected has error, got = %+v", got)
				}
			} else if err != nil {
				t.Errorf("ParseUrl() expected has not error, got = %s", err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUrl() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUrl_String(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		username string
		password string
		host     string
		path     string
		query    Query
		fragment string
		port     uint16
		want     string
	}{
		{
			name:   "Simple HTTP no path or query",
			scheme: "http",
			host:   "example.com",
			want:   "http://example.com",
		},
		{
			name:   "Path with spaces",
			scheme: "http",
			host:   "example.com",
			path:   "/foo bar/baz",
			want:   "http://example.com/foo%20bar/baz",
		},
		{
			name:   "Query with special characters",
			scheme: "http",
			host:   "example.com",
			query:  Query{"a b": "c=d&"},
			want:   "http://example.com?a+b=c%3Dd%26",
		},
		{
			name:     "Full URL with all fields",
			scheme:   "https",
			username: "user",
			password: "pass",
			host:     "example.com",
			path:     "/search results",
			query:    Query{"q": "golang & rust", "lang": "en"},
			fragment: "section 1",
			port:     8443,
			want:     "https://user:pass@example.com:8443/search%20results?lang=en&q=golang+%26+rust#section%201",
		},
		{
			name:   "IPv6 with port and path",
			scheme: "http",
			host:   "[2001:db8::1]",
			port:   8080,
			path:   "/file/ü",
			want:   "http://[2001:db8::1]:8080/file/%C3%BC",
		},
		{
			name:   "Empty query key and value",
			scheme: "https",
			host:   "example.com",
			query:  Query{"": ""},
			want:   "https://example.com?=",
		},
		{
			name:     "Username only",
			scheme:   "http",
			host:     "example.com",
			username: "bob",
			want:     "http://bob@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &Url{
				Scheme:   tt.scheme,
				Username: tt.username,
				Password: tt.password,
				Host:     tt.host,
				Path:     tt.path,
				Query:    tt.query,
				Fragment: tt.fragment,
				Port:     tt.port,
			}
			if got := url.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
