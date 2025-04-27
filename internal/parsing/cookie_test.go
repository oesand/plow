package parsing

import (
	"github.com/oesand/giglet/specs"
	"maps"
	"reflect"
	"testing"
	"time"
)

func TestParseCookieHeader(t *testing.T) {
	tests := []struct {
		name       string
		cookieText string
		want       map[string]string
	}{
		{
			name:       "Single cookie",
			cookieText: "sessionId=abc123",
			want:       map[string]string{"sessionId": "abc123"},
		},
		{
			name:       "Multiple cookies",
			cookieText: "user=JohnDoe; sessionId=xyz789; theme=dark",
			want: map[string]string{
				"user":      "JohnDoe",
				"sessionId": "xyz789",
				"theme":     "dark",
			},
		},
		{
			name:       "Cookies with whitespace",
			cookieText: "foo = bar ; baz = qux",
			want: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			name:       "Cookie with empty value",
			cookieText: "token=; user=guest",
			want: map[string]string{
				"token": "",
				"user":  "guest",
			},
		},
		{
			name:       "Cookie with equal sign in value",
			cookieText: "auth=abc=123; theme=light",
			want: map[string]string{
				"auth":  "abc=123",
				"theme": "light",
			},
		},
		{
			name:       "Empty cookie string",
			cookieText: "",
			want:       map[string]string{},
		},
		{
			name:       "Trailing semicolon",
			cookieText: "user=JohnDoe;",
			want:       map[string]string{"user": "JohnDoe"},
		},
		{
			name:       "Multiple semicolons and empty entries",
			cookieText: "foo=bar;;baz=qux;;;",
			want: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			name:       "Cookie with spaces around key and value",
			cookieText: " name = value ;  test = 123 ",
			want: map[string]string{
				"name": "value",
				"test": "123",
			},
		},
		{
			name:       "Cookie with special characters",
			cookieText: "token=%21%40%23; path=/; HttpOnly",
			want: map[string]string{
				"token":    "%21%40%23",
				"path":     "/",
				"HttpOnly": "",
			},
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			itr := ParseCookieHeader(data.cookieText)
			actual := maps.Collect(itr)
			if !reflect.DeepEqual(actual, data.want) {
				t.Errorf("ParseCookieHeader() = %v, want %v", actual, data.want)
			}
		})
	}
}

func TestParseSetCookieHeader(t *testing.T) {
	time, _ := time.Parse(specs.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
	tests := []struct {
		name       string
		cookieText string
		want       *specs.Cookie
	}{
		{
			name:       "Basic key-value",
			cookieText: "key=value",
			want: &specs.Cookie{
				Name:  "key",
				Value: "value",
			},
		},
		{
			name:       "With path and domain",
			cookieText: "key=value; Path=/; Domain=example.com",
			want: &specs.Cookie{
				Name:   "key",
				Value:  "value",
				Path:   "/",
				Domain: "example.com",
			},
		},
		{
			name:       "Secure and HttpOnly flags",
			cookieText: "key=value; Secure; HttpOnly",
			want: &specs.Cookie{
				Name:     "key",
				Value:    "value",
				Secure:   true,
				HttpOnly: true,
			},
		},
		{
			name:       "With Max-Age",
			cookieText: "key=value; Max-Age=3600",
			want: &specs.Cookie{
				Name:   "key",
				Value:  "value",
				MaxAge: 3600,
			},
		},
		{
			name:       "With Expires",
			cookieText: "key=value; Expires=Wed, 21 Oct 2015 07:28:00 GMT",
			want: &specs.Cookie{
				Name:    "key",
				Value:   "value",
				Expires: time,
			},
		},
		{
			name:       "SameSite attribute",
			cookieText: "key=value; SameSite=Strict",
			want: &specs.Cookie{
				Name:     "key",
				Value:    "value",
				SameSite: "Strict",
			},
		},
		{
			name:       "Mixed all attributes",
			cookieText: "key=value; Path=/; Domain=example.com; Max-Age=60; Secure; HttpOnly; SameSite=Lax",
			want: &specs.Cookie{
				Name:     "key",
				Value:    "value",
				Path:     "/",
				Domain:   "example.com",
				MaxAge:   60,
				Secure:   true,
				HttpOnly: true,
				SameSite: "Lax",
			},
		},
	}

	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			if got := ParseSetCookieHeader(data.cookieText); !reflect.DeepEqual(got, data.want) {
				t.Errorf("ParseSetCookieHeader() = %v, want %v", got, data.want)
			}
		})
	}
}
