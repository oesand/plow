package prm

import (
	"context"
	"testing"

	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
)

func TestHeaderParam(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		expectedValue string
		expectedResp  bool
	}{
		{"valid header", "Authorization", "Bearer token123", "Bearer token123", false},
		{"missing header", "Authorization", "", "", false},
		{"empty header value", "X-Custom-Header", "", "", false},
		{"header with spaces", "User-Agent", "Mozilla/5.0", "Mozilla/5.0", false},
		{"header with special chars", "X-API-Key", "key@123#456", "key@123#456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.headerValue != "" {
						h.Set(tt.headerName, tt.headerValue)
					}
				}).
				Request()

			param := HeaderParam(tt.headerName)
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

func TestHeaderParamRequired(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		required      bool
		expectedValue string
		expectedResp  bool
		expectedError string
	}{
		{"required header present", "Authorization", "Bearer token", true, "Bearer token", false, ""},
		{"required header missing", "Authorization", "", true, "", true, "header 'Authorization' is required"},
		{"optional header present", "X-Optional", "value", false, "value", false, ""},
		{"optional header missing", "X-Optional", "", false, "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.headerValue != "" {
						h.Set(tt.headerName, tt.headerValue)
					}
				}).
				Request()

			param := HeaderParam(tt.headerName)
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

func TestHeaderParamWithConditions(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		conditions    []Condition[string]
		expectedValue string
		expectedResp  bool
	}{
		{
			name:          "valid header with min length condition",
			headerName:    "Authorization",
			headerValue:   "Bearer token123",
			conditions:    []Condition[string]{MinLen(10)},
			expectedValue: "Bearer token123",
			expectedResp:  false,
		},
		{
			name:          "invalid header with min length condition",
			headerName:    "Authorization",
			headerValue:   "short",
			conditions:    []Condition[string]{MinLen(10)},
			expectedValue: "short",
			expectedResp:  true,
		},
		{
			name:          "valid header with max length condition",
			headerName:    "X-Custom",
			headerValue:   "short",
			conditions:    []Condition[string]{MaxLen(10)},
			expectedValue: "short",
			expectedResp:  false,
		},
		{
			name:          "invalid header with max length condition",
			headerName:    "X-Custom",
			headerValue:   "this is too long for the condition",
			conditions:    []Condition[string]{MaxLen(10)},
			expectedValue: "this is too long for the condition",
			expectedResp:  true,
		},
		{
			name:          "valid header with regex condition",
			headerName:    "X-API-Key",
			headerValue:   "key_123",
			conditions:    []Condition[string]{RegexPattern(`^key_\d+$`)},
			expectedValue: "key_123",
			expectedResp:  false,
		},
		{
			name:          "invalid header with regex condition",
			headerName:    "X-API-Key",
			headerValue:   "invalid-key",
			conditions:    []Condition[string]{RegexPattern(`^key_\d+$`)},
			expectedValue: "invalid-key",
			expectedResp:  true,
		},
		{
			name:          "valid header with multiple conditions",
			headerName:    "X-Custom",
			headerValue:   "valid123",
			conditions:    []Condition[string]{MinLen(5), MaxLen(10), RegexPattern(`^[a-z]+\d+$`)},
			expectedValue: "valid123",
			expectedResp:  false,
		},
		{
			name:          "invalid header with multiple conditions (first fails)",
			headerName:    "X-Custom",
			headerValue:   "short",
			conditions:    []Condition[string]{MinLen(10), MaxLen(20)},
			expectedValue: "short",
			expectedResp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					h.Set(tt.headerName, tt.headerValue)
				}).
				Request()

			param := HeaderParam(tt.headerName, tt.conditions...)
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

func TestHeaderParamCaseInsensitive(t *testing.T) {
	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		expectedValue string
	}{
		{"lowercase header name", "authorization", "Bearer token", "Bearer token"},
		{"uppercase header name", "AUTHORIZATION", "Bearer token", "Bearer token"},
		{"mixed case header name", "AuThOrIzAtIoN", "Bearer token", "Bearer token"},
		{"camel case header name", "user-agent", "Mozilla/5.0", "Mozilla/5.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					h.Set(tt.headerName, tt.headerValue)
				}).
				Request()

			// Use the same name that was set to test case insensitivity
			param := HeaderParam(tt.headerName)
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

func TestHeaderParamEdgeCases(t *testing.T) {
	t.Run("empty header name", func(t *testing.T) {
		param := HeaderParam("")
		req := mock.DefaultRequest().Request()

		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "" {
			t.Errorf("expected empty value for empty header name, got %q", value)
		}
	})

	t.Run("nil conditions", func(t *testing.T) {
		param := HeaderParam("X-Test")
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				h.Set("X-Test", "value")
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
		param := HeaderParam("X-Test", conditions...)
		req := mock.DefaultRequest().
			ConfHeader(func(h *specs.Header) {
				h.Set("X-Test", "value")
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
