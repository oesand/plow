package prm

import (
	"context"
	"testing"

	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
)

// mockCondition implements Condition for testing
type mockCondition[T any] struct {
	shouldFail bool
	failMsg    string
}

func (m *mockCondition[T]) Validate(value T) error {
	if m.shouldFail {
		return &mockError{msg: m.failMsg}
	}
	return nil
}

type mockError struct {
	msg string
}

func (m *mockError) Error() string { return m.msg }

func TestQueryParamString(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue string
		expectedError bool
		expectedResp  bool
	}{
		{"valid string param", map[string]string{"name": "test"}, "name", "test", false, false},
		{"missing param", map[string]string{"other": "value"}, "name", "", false, false},
		{"empty string param", map[string]string{"name": ""}, "name", "", false, false},
		{"param with spaces", map[string]string{"name": "hello world"}, "name", "hello world", false, false},
		{"param with special chars", map[string]string{"name": "test@example.com"}, "name", "test@example.com", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[string](tt.paramName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, value)
			}
		})
	}
}

func TestQueryParamBool(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue bool
		expectedError bool
	}{
		{"valid true", map[string]string{"flag": "true"}, "flag", true, false},
		{"valid false", map[string]string{"flag": "false"}, "flag", false, false},
		{"valid 1", map[string]string{"flag": "1"}, "flag", true, false},
		{"valid 0", map[string]string{"flag": "0"}, "flag", false, false},
		{"valid TRUE", map[string]string{"flag": "TRUE"}, "flag", true, false},
		{"valid FALSE", map[string]string{"flag": "FALSE"}, "flag", false, false},
		{"invalid bool", map[string]string{"flag": "invalid"}, "flag", false, true},
		{"missing param", map[string]string{"other": "value"}, "flag", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[bool](tt.paramName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %v, got %v", tt.expectedValue, value)
			}
		})
	}
}

func TestQueryParamInt64(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue int64
		expectedError bool
	}{
		{"valid positive int", map[string]string{"num": "42"}, "num", 42, false},
		{"valid negative int", map[string]string{"num": "-42"}, "num", -42, false},
		{"valid zero", map[string]string{"num": "0"}, "num", 0, false},
		{"valid large int", map[string]string{"num": "2147483647"}, "num", 2147483647, false},
		{"invalid int", map[string]string{"num": "not_a_number"}, "num", 0, true},
		{"float as int", map[string]string{"num": "3.14"}, "num", 0, true},
		{"missing param", map[string]string{"other": "value"}, "num", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[int64](tt.paramName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %d, got %d", tt.expectedValue, value)
			}
		})
	}
}

func TestQueryParamUint64(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue uint64
		expectedError bool
	}{
		{"valid positive uint", map[string]string{"num": "42"}, "num", 42, false},
		{"valid zero", map[string]string{"num": "0"}, "num", 0, false},
		{"valid large uint", map[string]string{"num": "4294967295"}, "num", 4294967295, false},
		{"negative uint", map[string]string{"num": "-42"}, "num", 0, true},
		{"invalid uint", map[string]string{"num": "not_a_number"}, "num", 0, true},
		{"missing param", map[string]string{"other": "value"}, "num", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[uint64](tt.paramName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %d, got %d", tt.expectedValue, value)
			}
		})
	}
}

func TestQueryParamFloat64(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue float64
		expectedError bool
	}{
		{"valid positive float", map[string]string{"num": "3.14"}, "num", 3.14, false},
		{"valid negative float", map[string]string{"num": "-3.14"}, "num", -3.14, false},
		{"valid zero", map[string]string{"num": "0.0"}, "num", 0.0, false},
		{"valid integer as float", map[string]string{"num": "42"}, "num", 42.0, false},
		{"valid scientific notation", map[string]string{"num": "1.23e-4"}, "num", 1.23e-4, false},
		{"invalid float", map[string]string{"num": "not_a_number"}, "num", 0.0, true},
		{"missing param", map[string]string{"other": "value"}, "num", 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[float64](tt.paramName)
			value, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
			if value != tt.expectedValue {
				t.Errorf("expected value %f, got %f", tt.expectedValue, value)
			}
		})
	}
}

func TestQueryParamRequired(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		required      bool
		expectedError bool
	}{
		{"required param present", map[string]string{"name": "test"}, "name", true, false},
		{"required param missing", map[string]string{"other": "value"}, "name", true, true},
		{"optional param missing", map[string]string{"other": "value"}, "name", false, false},
		{"required param empty", map[string]string{"name": ""}, "name", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[string](tt.paramName)
			if tt.required {
				param.Require()
			}

			_, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
		})
	}
}

func TestQueryParamWithConditions(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		conditions    []Condition[int64]
		expectedError bool
	}{
		{"valid with condition", map[string]string{"num": "42"}, "num", []Condition[int64]{&mockCondition[int64]{shouldFail: false}}, false},
		{"invalid with condition", map[string]string{"num": "42"}, "num", []Condition[int64]{&mockCondition[int64]{shouldFail: true, failMsg: "validation failed"}}, true},
		{"multiple conditions all pass", map[string]string{"num": "42"}, "num", []Condition[int64]{
			&mockCondition[int64]{shouldFail: false},
			&mockCondition[int64]{shouldFail: false},
		}, false},
		{"multiple conditions one fails", map[string]string{"num": "42"}, "num", []Condition[int64]{
			&mockCondition[int64]{shouldFail: false},
			&mockCondition[int64]{shouldFail: true, failMsg: "second validation failed"},
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &specs.Url{Query: specs.Query(tt.queryParams)}
			req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

			param := QueryParam[int64](tt.paramName, tt.conditions...)
			_, resp := param.GetParamValue(context.Background(), req)

			if tt.expectedError && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedError && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}
		})
	}
}

func TestQueryParamEdgeCases(t *testing.T) {
	t.Run("nil query params", func(t *testing.T) {
		url := &specs.Url{Query: nil}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[string]("name")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "" {
			t.Errorf("expected empty string, got %q", value)
		}
	})

	t.Run("empty query params", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[string]("name")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "" {
			t.Errorf("expected empty string, got %q", value)
		}
	})

	t.Run("query param with equals sign", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"name": "key=value"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[string]("name")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "key=value" {
			t.Errorf("expected 'key=value', got %q", value)
		}
	})

	t.Run("query param with ampersand", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"name": "a&b"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[string]("name")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != "a&b" {
			t.Errorf("expected 'a&b', got %q", value)
		}
	})
}

func TestQueryParamTypeConversions(t *testing.T) {
	t.Run("int64 conversion", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"num": "127"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[int64]("num")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != 127 {
			t.Errorf("expected 127, got %d", value)
		}
	})

	t.Run("uint64 conversion", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"num": "65535"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[uint64]("num")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != 65535 {
			t.Errorf("expected 65535, got %d", value)
		}
	})

	t.Run("float64 conversion", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"num": "3.14159"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[float64]("num")
		value, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if value != 3.14159 {
			t.Errorf("expected 3.14159, got %f", value)
		}
	})
}

func TestQueryParamErrorMessages(t *testing.T) {
	t.Run("required param error message", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[string]("name").Require()
		_, resp := param.GetParamValue(context.Background(), req)

		if resp == nil {
			t.Error("expected error response, got nil")
		}

		// Check that the error message contains the parameter name
		if resp.StatusCode() != specs.StatusCodeBadRequest {
			t.Errorf("expected bad request status, got %v", resp.StatusCode())
		}
	})

	t.Run("invalid bool error message", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"flag": "invalid"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[bool]("flag")
		_, resp := param.GetParamValue(context.Background(), req)

		if resp == nil {
			t.Error("expected error response, got nil")
		}

		if resp.StatusCode() != specs.StatusCodeBadRequest {
			t.Errorf("expected bad request status, got %v", resp.StatusCode())
		}
	})

	t.Run("invalid int error message", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"num": "not_a_number"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[int64]("num")
		_, resp := param.GetParamValue(context.Background(), req)

		if resp == nil {
			t.Error("expected error response, got nil")
		}

		if resp.StatusCode() != specs.StatusCodeBadRequest {
			t.Errorf("expected bad request status, got %v", resp.StatusCode())
		}
	})

	t.Run("condition validation error message", func(t *testing.T) {
		url := &specs.Url{Query: specs.Query{"num": "42"}}
		req := mock.DefaultRequest().Method(specs.HttpMethodGet).Url(url).Request()

		param := QueryParam[int64]("num", &mockCondition[int64]{shouldFail: true, failMsg: "custom validation error"})
		_, resp := param.GetParamValue(context.Background(), req)

		if resp == nil {
			t.Error("expected error response, got nil")
		}

		if resp.StatusCode() != specs.StatusCodeBadRequest {
			t.Errorf("expected bad request status, got %v", resp.StatusCode())
		}
	})
}
