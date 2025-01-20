package giglet

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/oesand/giglet/specs"
	"io"
	"iter"
	"net"
	"strconv"
	"strings"
	"time"
)

type Handler func(request Request) Response
type HijackHandler func(conn net.Conn)
type NextProtoHandler func(conn *tls.Conn)
type ConnHandler func(addr net.Conn, context context.Context) context.Context
type EventHandler func()

var (
	DefaultServerName                = "giglet"
	HeadlineMaxLength          int64 = 512
	DefaultContentMaxSizeBytes int64 = 5 << 20 // 5 MB

	ErrorCancelled = &specs.GigletError{Err: errors.New("cancelled")}

	zeroDialer         net.Dialer
	zeroTime           time.Time
	httpV1NextProtoTLS = "http/1.1"

	httpVersionPrefix = []byte("HTTP/")
	httpV10           = []byte("HTTP/1.0")
	httpV11           = []byte("HTTP/1.1")
	httpV2            = []byte("HTTP/2.0")

	directCrlf  = []byte("\r\n")
	directColon = []byte(": ")
	emptyBytes  = []byte("")

	rawCloseHeaders             = []byte("Content-Type: text/plain; charset=utf-8\r\nConnection: close\r\n")
	responseDowngradeHTTPS      = []byte("HTTP/1.0 400 Bad Request\r\n\r\nSent an HTTP request to an HTTPS server.\n")
	responseNotProcessableError = []byte("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/plain; charset=utf-8\r\nConnection: close\r\n\r\n500 Unknown error while processing the request\n")
	responseUnsupportedEncoding = []byte("HTTP/1.1 501 Not Implemented\r\nContent-Type: text/plain; charset=utf-8\r\nConnection: close\r\n\r\n501 Unsupported transfer encoding\n")
)

func validationErr(err string, a ...any) error {
	return &specs.GigletError{
		Op:  "validation",
		Err: fmt.Errorf(err, a...),
	}
}

type statusErrorResponse struct {
	code specs.StatusCode
	text string
}

func (err *statusErrorResponse) Error() string {
	return string(err.code.Detail()) + ": " + err.text
}

func (err *statusErrorResponse) Write(writer io.Writer) {
	var buf bytes.Buffer

	buf.Write(httpV11)
	buf.WriteByte(' ')
	buf.Write(err.code.Detail())
	buf.Write(directCrlf)
	buf.Write(rawCloseHeaders)
	buf.Write(directCrlf)
	buf.WriteString(err.text)

	buf.WriteTo(writer)
}

func parseCookieHeader(cookie string) iter.Seq2[string, string] {
	splitted := strings.Split(cookie, "; ")

	return func(yield func(string, string) bool) {
		for _, pair := range splitted {
			key, value, ok := strings.Cut(pair, "=")
			if !ok || len(key) == 0 || len(value) == 0 {
				continue
			}
			if !yield(key, value) {
				break
			}
		}
	}
}

func parseSetCookieHeader(cookieValue string) *specs.Cookie {
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
						cookie.Expires = zeroTime
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
