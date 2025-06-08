package specs

import (
	"testing"
	"time"
)

func makeCookie(name string, expired bool) Cookie {
	c := Cookie{
		Name:  name,
		Value: "value_" + name,
		Path:  "/",
	}

	if expired {
		c.Expires = time.Now().Add(-time.Hour)
	} else {
		c.Expires = time.Now().Add(time.Hour)
	}

	return c
}

func TestSetAndGetCookie(t *testing.T) {
	jar := &CookieJar{}
	c := makeCookie("session", false)

	jar.SetCookie("www.example.com", c)

	got := jar.GetCookie("www.example.com", "session")
	if got == nil {
		t.Fatal("expected cookie, got nil")
	}
	if got.Value != c.Value {
		t.Errorf("expected value %s, got %s", c.Value, got.Value)
	}
}

func TestGetCookieNilMap(t *testing.T) {
	jar := &CookieJar{} // cookies map is nil

	result := jar.GetCookie("example.com", "anything")
	if result != nil {
		t.Error("Expected nil due to uninitialized cookie map")
	}
}

func TestGetCookieExpired(t *testing.T) {
	jar := &CookieJar{}
	expired := makeCookie("expired", true)

	jar.SetCookie("www.example.com", expired)

	got := jar.GetCookie("www.example.com", "expired")
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestCookiesNilOrEmpty(t *testing.T) {
	jar := &CookieJar{}
	seq := jar.Cookies("example.com")

	count := 0
	for _ = range seq {
		count++
	}
	if count != 0 {
		t.Errorf("Expected 0 cookies, got %d", count)
	}
}

func TestGetCookieWrongDomain(t *testing.T) {
	jar := &CookieJar{}
	c := makeCookie("token", false)

	jar.SetCookie("example.com", c)
	result := jar.GetCookie("other.com", "token")
	if result != nil {
		t.Errorf("Expected nil for mismatched domain, got: %+v", result)
	}
}

func TestCookiesEmptyForUnknownHost(t *testing.T) {
	jar := &CookieJar{}
	c := makeCookie("session", false)
	jar.SetCookie("example.com", c)

	count := 0
	for _ = range jar.Cookies("other.com") {
		count++
	}
	if count != 0 {
		t.Errorf("Expected 0 cookies, got %d", count)
	}
}

func TestSetCookies(t *testing.T) {
	jar := &CookieJar{}
	c1 := makeCookie("cookie1", false)
	c2 := makeCookie("cookie2", false)

	jar.SetCookies("example.com", []Cookie{c1, c2})

	if jar.GetCookie("example.com", "cookie1") == nil {
		t.Error("cookie1 not found")
	}
	if jar.GetCookie("example.com", "cookie2") == nil {
		t.Error("cookie2 not found")
	}
}

func TestCookiesIter(t *testing.T) {
	jar := &CookieJar{}
	c1 := makeCookie("a", false)
	c2 := makeCookie("b", true) // expired
	c3 := makeCookie("c", false)

	jar.SetCookies("example.com", []Cookie{c1, c2, c3})

	collected := []string{}
	for c := range jar.Cookies("example.com") {
		collected = append(collected, c.Name)
	}

	if len(collected) != 2 {
		t.Errorf("expected 2 cookies, got %d: %v", len(collected), collected)
	}
}

func TestSetCookiesIterEarlyStop(t *testing.T) {
	jar := &CookieJar{}
	c1 := makeCookie("a", false)
	c2 := makeCookie("b", false)

	jar.SetCookiesIter("example.com", func(yield func(Cookie) bool) {
		yield(c1)
		// Simulate early stop
		yield(c2)
	})

	if jar.GetCookie("example.com", "a") == nil {
		t.Error("cookie a missing")
	}
	if jar.GetCookie("example.com", "b") == nil {
		t.Error("cookie b missing")
	}
}

func TestSetCookiesIterAllExpired(t *testing.T) {
	jar := &CookieJar{}
	expired := makeCookie("expired1", true)
	expired2 := makeCookie("expired2", true)

	jar.SetCookiesIter("example.com", func(yield func(Cookie) bool) {
		yield(expired)
		yield(expired2)
	})

	if jar.GetCookie("example.com", "expired1") != nil {
		t.Error("Expected expired1 to be discarded")
	}
	if jar.GetCookie("example.com", "expired2") != nil {
		t.Error("Expected expired2 to be discarded")
	}
}

func TestIsExpiredWithMaxAge(t *testing.T) {
	c := Cookie{
		MaxAge: 10, // seconds
	}

	now := time.Now()
	expired := c.IsExpired(now)

	if expired {
		t.Error("cookie should not be expired")
	}
	if c.Expires.Before(now) {
		t.Error("Expires not updated correctly")
	}
}
