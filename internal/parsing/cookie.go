package parsing

import (
	"github.com/oesand/giglet/specs"
	"iter"
	"strconv"
	"strings"
	"time"
)

func ParseCookieHeader(cookieText string) iter.Seq2[string, string] {
	pairs := strings.Split(cookieText, ";")

	return func(yield func(string, string) bool) {
		for _, part := range pairs {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			pair := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(pair[0])
			value := ""
			if len(pair) > 1 {
				value = strings.TrimSpace(pair[1])
			}

			if !yield(key, value) {
				break
			}
		}
	}
}

func ParseSetCookieHeader(cookieValue string) *specs.Cookie {
	if len(cookieValue) == 0 {
		return nil
	}
	lastIndex := len(cookieValue) - 1
	var cookie *specs.Cookie
	var key string
	var curr int
	for i, c := range cookieValue {
		if cookie == nil {
			if key == "" {
				switch c {
				case ';':
					break
				case '=':
					key = cookieValue[:i]
					curr = i + 1
				}
			} else if c == ';' || i == lastIndex {
				if i-curr == 0 {
					break
				}
				var value string
				if i == lastIndex {
					value = cookieValue[curr:]
				} else {
					value = cookieValue[curr:i]
				}
				cookie = &specs.Cookie{
					Name:  key,
					Value: value,
				}
				key = ""
				curr = i + 2
			}
		} else if i > curr {
			if key == "" && c == '=' {
				key = cookieValue[curr:i]
				if key == "" || !(strings.EqualFold(key, "Expires") ||
					strings.EqualFold(key, "Max-Age") ||
					strings.EqualFold(key, "Domain") ||
					strings.EqualFold(key, "Path") ||
					strings.EqualFold(key, "SameSite")) {
					key = ""
				} else {
					curr = i + 1
				}
			} else if c == ';' || i == lastIndex {
				var value string
				if i == lastIndex {
					value = cookieValue[curr:]
				} else {
					value = cookieValue[curr:i]
				}
				if key == "" {
					switch {
					case strings.EqualFold(value, "Secure"):
						cookie.Secure = true
					case strings.EqualFold(value, "HttpOnly"):
						cookie.HttpOnly = true
					}
				} else {
					switch {
					case strings.EqualFold(key, "Expires") && cookie.MaxAge <= 0:
						cookie.Expires, _ = time.Parse(specs.TimeFormat, value)
					case strings.EqualFold(key, "Max-Age"):
						cookie.Expires = time.Time{}
						cookie.MaxAge, _ = strconv.ParseUint(value, 10, 64)
					case strings.EqualFold(key, "Domain"):
						cookie.Domain = value
					case strings.EqualFold(key, "Path"):
						cookie.Path = value
					case strings.EqualFold(key, "SameSite"):
						switch {
						case strings.EqualFold(value, "None"):
							cookie.SameSite = specs.CookieSameSiteNoneMode
						case strings.EqualFold(value, "Lax"):
							cookie.SameSite = specs.CookieSameSiteLaxMode
						case strings.EqualFold(value, "Strict"):
							cookie.SameSite = specs.CookieSameSiteStrictMode
						}
					}
				}
				key = ""
				curr = i + 2
			}

			switch c {
			case '=':

				switch {
				case strings.EqualFold(key, "Expires"),
					strings.EqualFold(key, "Max-Age"),
					strings.EqualFold(key, "Domain"),
					strings.EqualFold(key, "Path"),
					strings.EqualFold(key, "Path"),
					strings.EqualFold(key, "SameSite"):
					curr = i + 1
				default:
				}

			case ';':

			}
		}
	}
	return cookie
}
