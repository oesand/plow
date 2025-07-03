package server

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"strings"
	"testing"
	"time"
)

func TestWriteResponseHead(t *testing.T) {
	tests := []struct {
		name     string
		is11     bool
		code     specs.StatusCode
		header   *specs.Header
		expected string
	}{
		{
			name: "HTTP/1.1 OK with headers and cookies",
			is11: true,
			code: specs.StatusCodeOK,
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Header", "Value")
				header.Set("Content-Type", "text/html")
				header.SetCookieValue("Cookie", "Value")
				header.SetCookie(specs.Cookie{
					Name:     "sessionid",
					Value:    "abc123",
					Domain:   "example.com",
					MaxAge:   3600,
					Expires:  time.Date(2025, time.April, 1, 12, 0, 0, 0, time.UTC),
					Path:     "/home",
					HttpOnly: true,
					Secure:   true,
					SameSite: specs.CookieSameSiteStrictMode,
				})
			}),
			expected: strings.Join([]string{
				"HTTP/1.1 200 OK",
				"Content-Type: text/html",
				"Header: Value",
				"Set-Cookie: Cookie=Value",
				"Set-Cookie: sessionid=abc123; Max-Age=3600; Domain=example.com; Path=/home; HttpOnly; Secure; SameSite=Strict",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name: "HTTP/1.0 with status code and cookie",
			is11: false,
			code: specs.StatusCodeNotFound,
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Content-Type", "text/html")
				header.SetCookieValue("sessionid", "xyz123")
			}),
			expected: strings.Join([]string{
				"HTTP/1.0 404 Not Found",
				"Content-Type: text/html",
				"Set-Cookie: sessionid=xyz123",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:     "HTTP/1.1 without any headers or cookies",
			is11:     true,
			code:     specs.StatusCodeOK,
			header:   specs.NewHeader(),
			expected: "HTTP/1.1 200 OK\r\n\r\n",
		},
		{
			name: "HTTP/1.1 with multiple cookies",
			is11: true,
			code: specs.StatusCodeForbidden,
			header: specs.NewHeader(func(header *specs.Header) {
				header.SetCookieValue("user", "john123")
				header.SetCookieValue("auth", "token456")
			}),
			expected: strings.Join([]string{
				"HTTP/1.1 403 Forbidden",
				"Set-Cookie: auth=token456",
				"Set-Cookie: user=john123",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name: "HTTP/1.1 with empty cookie values",
			is11: true,
			code: specs.StatusCodeOK,
			header: specs.NewHeader(func(header *specs.Header) {
				header.SetCookie(specs.Cookie{
					Name: "emptycookie",
				})
			}),
			expected: strings.Join([]string{
				"HTTP/1.1 200 OK",
				"Set-Cookie: emptycookie=",
			}, "\r\n") + "\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			_, _ = WriteResponseHead(writer, tt.is11, tt.code, tt.header)
			if gotText := writer.String(); gotText != tt.expected {
				t.Errorf("WriteResponseHead() gotWriter = \n%v\nwant \n%v", gotText, tt.expected)
			}
		})
	}
}
