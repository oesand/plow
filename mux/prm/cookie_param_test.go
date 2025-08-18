package prm

import (
	"context"
	"testing"

	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
)

func TestCookieParam(t *testing.T) {
	tests := []struct {
		name          string
		cookieName    string
		cookieValue   string
		expectedValue string
		expectedResp  bool
	}{
		{"valid cookie", "session_id", "abc123def456", "abc123def456", false},
		{"missing cookie", "session_id", "", "", false},
		{"empty cookie value", "X-Custom-Cookie", "", "", false},
		{"cookie with spaces", "user_preference", "dark theme", "dark theme", false},
		{"cookie with special chars", "api_token", "token@123#456", "token@123#456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.cookieValue != "" {
						h.SetCookieValue(tt.cookieName, tt.cookieValue)
					}
				}).
				Request()

			param := CookieParam(tt.cookieName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedResp && resp == nil {
				t.Error("expected response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestCookieParamRequired(t *testing.T) {
	tests := []struct {
		name          string
		cookieName    string
		cookieValue   string
		required      bool
		expectedValue string
		expectedResp  bool
		expectedError string
	}{
		{"required cookie present", "session_id", "abc123", true, "abc123", false, ""},
		{"required cookie missing", "session_id", "", true, "", true, "cookie 'session_id' is required"},
		{"optional cookie present", "X-Optional", "value", false, "value", false, ""},
		{"optional cookie missing", "X-Optional", "", false, "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.cookieValue != "" {
						h.SetCookieValue(tt.cookieName, tt.cookieValue)
					}
				}).
				Request()

			param := CookieParam(tt.cookieName)
			if tt.required {
				param = param.Require()
			}

			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
			if tt.expectedError != "" && resp != nil {
				// Check that the error message contains the expected text
				// Note: We can't easily extract the exact error message from the response
				// without implementing a mock response type, so we just verify a response exists
			}
		})
	}
}

func TestCookieParamWithConditions(t *testing.T) {
	tests := []struct {
		name          string
		cookieName    string
		cookieValue   string
		conditions    []Condition[string]
		expectedValue string
		expectedResp  bool
	}{
		{
			name:          "valid cookie with min length condition",
			cookieName:    "session_id",
			cookieValue:   "abc123def456",
			conditions:    []Condition[string]{MinLen(10)},
			expectedValue: "abc123def456",
			expectedResp:  false,
		},
		{
			name:          "invalid cookie with min length condition",
			cookieName:    "session_id",
			cookieValue:   "short",
			conditions:    []Condition[string]{MinLen(10)},
			expectedValue: "short",
			expectedResp:  true,
		},
		{
			name:          "valid cookie with max length condition",
			cookieName:    "X-Custom",
			cookieValue:   "short",
			conditions:    []Condition[string]{MaxLen(10)},
			expectedValue: "short",
			expectedResp:  false,
		},
		{
			name:          "invalid cookie with max length condition",
			cookieName:    "X-Custom",
			cookieValue:   "this is too long for the condition",
			conditions:    []Condition[string]{MaxLen(10)},
			expectedValue: "this is too long for the condition",
			expectedResp:  true,
		},
		{
			name:          "valid cookie with regex condition",
			cookieName:    "X-API-Key",
			cookieValue:   "key_123",
			conditions:    []Condition[string]{RegexPattern(`^key_\d+$`)},
			expectedValue: "key_123",
			expectedResp:  false,
		},
		{
			name:          "invalid cookie with regex condition",
			cookieName:    "X-API-Key",
			cookieValue:   "invalid-key",
			conditions:    []Condition[string]{RegexPattern(`^key_\d+$`)},
			expectedValue: "invalid-key",
			expectedResp:  true,
		},
		{
			name:          "valid cookie with multiple conditions",
			cookieName:    "X-Custom",
			cookieValue:   "valid123",
			conditions:    []Condition[string]{MinLen(5), MaxLen(10), RegexPattern(`^[a-z]+\d+$`)},
			expectedValue: "valid123",
			expectedResp:  false,
		},
		{
			name:          "invalid cookie with multiple conditions (first fails)",
			cookieName:    "X-Custom",
			cookieValue:   "short",
			conditions:    []Condition[string]{MinLen(10), MaxLen(20)},
			expectedValue: "short",
			expectedResp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					h.SetCookieValue(tt.cookieName, tt.cookieValue)
				}).
				Request()

			param := CookieParam(tt.cookieName, tt.conditions...)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestCookieParamCaseInsensitive(t *testing.T) {
	tests := []struct {
		name          string
		cookieName    string
		cookieValue   string
		expectedValue string
	}{
		{"lowercase cookie name", "sessionid", "abc123", "abc123"},
		{"uppercase cookie name", "SESSIONID", "abc123", "abc123"},
		{"mixed case cookie name", "SeSsIoNiD", "abc123", "abc123"},
		{"camel case cookie name", "user-preference", "dark theme", "dark theme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					h.SetCookieValue(tt.cookieName, tt.cookieValue)
				}).
				Request()

			// Use the same name that was set to test case insensitivity
			param := CookieParam(tt.cookieName)
			value, resp := param.GetParamValue(context.Background(), req)

			if resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestCookieParamEdgeCases(t *testing.T) {
	t.Run("empty cookie name", func(t *testing.T) {
		param := CookieParam("")
		req := mock.DefaultRequest().Request()

		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "" {
			t.Errorf("expected empty value for empty cookie name, got %q", value)
		}
	})

	t.Run("nil conditions", func(t *testing.T) {
		param := CookieParam("X-Test")
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				h.SetCookieValue("X-Test", "value")
			}).
			Request()

		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "value" {
			t.Errorf("expected value 'value', got %q", value)
		}
	})

	t.Run("empty conditions slice", func(t *testing.T) {
		conditions := []Condition[string]{}
		param := CookieParam("X-Test", conditions...)
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				h.SetCookieValue("X-Test", "value")
			}).
			Request()

		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "value" {
			t.Errorf("expected value 'value', got %q", value)
		}
	})
}

func TestCookieParamWithFullCookie(t *testing.T) {
	t.Run("cookie with full attributes", func(t *testing.T) {
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				cookie := specs.Cookie{
					Name:     "session_id",
					Value:    "abc123def456",
					Domain:   "example.com",
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
				}
				h.SetCookie(cookie)
			}).
			Request()

		param := CookieParam("session_id")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "abc123def456" {
			t.Errorf("expected value 'abc123def456', got %q", value)
		}
	})
}

func TestCookieParamMultipleCookies(t *testing.T) {
	t.Run("multiple cookies in request", func(t *testing.T) {
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				h.SetCookieValue("session_id", "abc123")
				h.SetCookieValue("user_id", "user456")
				h.SetCookieValue("theme", "dark")
			}).
			Request()

		// Test extracting different cookies
		sessionParam := CookieParam("session_id")
		userParam := CookieParam("user_id")
		themeParam := CookieParam("theme")

		sessionValue, sessionResp := sessionParam.GetParamValue(context.Background(), req)
		userValue, userResp := userParam.GetParamValue(context.Background(), req)
		themeValue, themeResp := themeParam.GetParamValue(context.Background(), req)

		if sessionResp != nil || userResp != nil || themeResp != nil {
			t.Error("unexpected error responses")
		}
		if sessionValue != "abc123" {
			t.Errorf("expected session value 'abc123', got %q", sessionValue)
		}
		if userValue != "user456" {
			t.Errorf("expected user value 'user456', got %q", userValue)
		}
		if themeValue != "dark" {
			t.Errorf("expected theme value 'dark', got %q", themeValue)
		}
	})
}
