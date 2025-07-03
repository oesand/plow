package parsing

import (
	"github.com/oesand/giglet/specs"
	"testing"
	"time"
)

func TestSetCookieBytes(t *testing.T) {
	tests := []struct {
		name   string
		cookie *specs.Cookie
		want   string
	}{
		{
			name: "Simple key-value",
			cookie: &specs.Cookie{
				Name:  "key",
				Value: "value",
			},
			want: "key=value",
		},
		{
			name: "Cookie with Path",
			cookie: &specs.Cookie{
				Name:  "sessionid",
				Value: "abc123",
				Path:  "/user",
			},
			want: "sessionid=abc123; Path=/user",
		},
		{
			name: "Cookie with Domain",
			cookie: &specs.Cookie{
				Name:   "token",
				Value:  "xyz",
				Domain: "example.com",
			},
			want: "token=xyz; Domain=example.com",
		},
		{
			name: "Cookie with MaxAge",
			cookie: &specs.Cookie{
				Name:   "remember",
				Value:  "true",
				MaxAge: 3600,
			},
			want: "remember=true; Max-Age=3600",
		},
		{
			name: "Cookie with Secure and HttpOnly",
			cookie: &specs.Cookie{
				Name:     "auth",
				Value:    "token",
				Secure:   true,
				HttpOnly: true,
			},
			want: "auth=token; HttpOnly; Secure",
		},
		{
			name: "Cookie with SameSite",
			cookie: &specs.Cookie{
				Name:     "samesite",
				Value:    "ok",
				SameSite: specs.CookieSameSiteStrictMode,
			},
			want: "samesite=ok; SameSite=Strict",
		},
		{
			name: "Cookie with Expires",
			cookie: &specs.Cookie{
				Name:    "expiring",
				Value:   "soon",
				Expires: time.Date(2025, time.April, 1, 12, 0, 0, 0, time.UTC),
			},
			want: "expiring=soon; Expires=Tue, 01 Apr 2025 12:00:00 GMT",
		},
		{
			name: "All parameters with Max-Age",
			cookie: &specs.Cookie{
				Name:     "user",
				Value:    "abcd1234",
				Path:     "/home",
				Domain:   "example.com",
				MaxAge:   3600,                                                // 1 hour
				Expires:  time.Date(2025, time.May, 1, 12, 0, 0, 0, time.UTC), // Expiration date
				Secure:   true,
				HttpOnly: true,
				SameSite: specs.CookieSameSiteLaxMode,
			},
			want: "user=abcd1234; Max-Age=3600; Domain=example.com; Path=/home; HttpOnly; Secure; SameSite=Lax",
		},
		{
			name: "All parameters without Max-Age",
			cookie: &specs.Cookie{
				Name:     "user",
				Value:    "abcd1234",
				Path:     "/home",
				Domain:   "example.com",                                       // 1 hour
				Expires:  time.Date(2025, time.May, 1, 12, 0, 0, 0, time.UTC), // Expiration date
				Secure:   true,
				HttpOnly: true,
				SameSite: specs.CookieSameSiteLaxMode,
			},
			want: "user=abcd1234; Expires=Thu, 01 May 2025 12:00:00 GMT; Domain=example.com; Path=/home; HttpOnly; Secure; SameSite=Lax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetCookieBytes(tt.cookie); tt.want != string(got) {
				t.Errorf("SetCookieBytes() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
