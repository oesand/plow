package writing

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"strings"
	"testing"
)

func TestWriteRequestHead(t *testing.T) {
	tests := []struct {
		name     string
		method   specs.HttpMethod
		url      *specs.Url
		header   *specs.Header
		expected string
	}{
		{
			name:   "Only url",
			method: specs.HttpMethodPost,
			url:    specs.MustParseUrl("/api/v1/resource"),
			header: specs.NewHeader(),
			expected: strings.Join([]string{
				"POST /api/v1/resource HTTP/1.1",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:   "Only url with query",
			method: specs.HttpMethodPost,
			url:    specs.MustParseUrl("/user?id=120"),
			header: specs.NewHeader(),
			expected: strings.Join([]string{
				"POST /user?id=120 HTTP/1.1",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:   "Only Cookies",
			method: specs.HttpMethodGet,
			url:    specs.MustParseUrl("/only-cookie"),
			header: specs.NewHeader(func(header *specs.Header) {
				header.SetCookieValue("session_id", "abc123")
				header.SetCookieValue("user_id", "4049")
			}),
			expected: strings.Join([]string{
				"GET /only-cookie HTTP/1.1",
				"Cookie: session_id=abc123; user_id=4049",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:   "Only headers",
			method: specs.HttpMethodPut,
			url:    specs.MustParseUrl("/update"),
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Content-Type", "application/json")
				header.Set("Authorization", "Bearer token")
			}),
			expected: strings.Join([]string{
				"PUT /update HTTP/1.1",
				"Content-Type: application/json",
				"Authorization: Bearer token",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:   "Empty URL path",
			method: specs.HttpMethodGet,
			url:    specs.MustParseUrl("/"),
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Test", "Value")
			}),
			expected: strings.Join([]string{
				"GET / HTTP/1.1",
				"Test: Value",
			}, "\r\n") + "\r\n\r\n",
		},
		{
			name:   "All",
			method: specs.HttpMethodPut,
			url:    specs.MustParseUrl("/all?one=two&three=four"),
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Content-Type", "application/json")
				header.Set("Authorization", "Bearer token")
				header.SetCookieValue("session_id", "abc123")
				header.SetCookieValue("user_id", "4049")
			}),
			expected: strings.Join([]string{
				"PUT /all?one=two&three=four HTTP/1.1",
				"Content-Type: application/json",
				"Authorization: Bearer token",
				"Cookie: session_id=abc123; user_id=4049",
			}, "\r\n") + "\r\n\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			_, _ = WriteRequestHead(writer, tt.method, tt.url, tt.header)
			if gotText := writer.String(); tt.expected != gotText {
				t.Errorf("WriteRequestHead() got text = \n%v\nexpected:\n%v", gotText, tt.expected)
			}
		})
	}
}
