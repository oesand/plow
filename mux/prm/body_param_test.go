package prm

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
)

func TestFormParam(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		contentType   string
		expectedData  specs.Query
		expectedResp  bool
		expectedError string
	}{
		{
			name:         "valid form data",
			body:         "username=john&email=john@example.com&age=25",
			contentType:  "application/x-www-form-urlencoded",
			expectedData: specs.Query{"username": "john", "email": "john@example.com", "age": "25"},
			expectedResp: false,
		},
		{
			name:          "empty form data",
			body:          "",
			contentType:   "application/x-www-form-urlencoded",
			expectedData:  nil,
			expectedResp:  true,
			expectedError: "request body is required",
		},
		{
			name:         "form data with special chars",
			body:         "name=John%20Doe&city=New%20York",
			contentType:  "application/x-www-form-urlencoded",
			expectedData: specs.Query{"name": "John Doe", "city": "New York"},
			expectedResp: false,
		},
		{
			name:         "form data with empty values",
			body:         "username=john&password=&email=john@example.com",
			contentType:  "application/x-www-form-urlencoded",
			expectedData: specs.Query{"username": "john", "password": "", "email": "john@example.com"},
			expectedResp: false,
		},
		{
			name:          "missing body",
			body:          "",
			contentType:   "",
			expectedData:  nil,
			expectedResp:  true,
			expectedError: "request body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.contentType != "" {
						h.Set("Content-Type", tt.contentType)
					}
				})

			// Only set body if there's actual content
			if tt.body != "" {
				req = req.Body(io.NopCloser(strings.NewReader(tt.body)))
			}

			request := req.Request()

			param := FormParam()
			result, resp := param.GetParamValue(context.Background(), request)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}

			if !tt.expectedResp {
				if result == nil && tt.expectedData != nil {
					t.Error("expected form data, got nil")
				}
				if result != nil && tt.expectedData == nil {
					t.Error("expected nil form data, got data")
				}
				if result != nil && tt.expectedData != nil {
					if len(result) != len(tt.expectedData) {
						t.Errorf("expected %d form fields, got %d", len(tt.expectedData), len(result))
					}
					for key, expectedValue := range tt.expectedData {
						if actualValue := result[key]; actualValue != expectedValue {
							t.Errorf("expected value %q for key %q, got %q", expectedValue, key, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestMultipartFormParam(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		boundary      string
		expectedResp  bool
		expectedError string
	}{
		{
			name:         "valid multipart form",
			body:         "--boundary123\r\nContent-Disposition: form-data; name=\"username\"\r\n\r\njohn\r\n--boundary123--",
			boundary:     "boundary123",
			expectedResp: false,
		},
		{
			name:         "multipart form with file",
			body:         "--boundary123\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\nContent-Type: text/plain\r\n\r\nfile content\r\n--boundary123--",
			boundary:     "boundary123",
			expectedResp: false,
		},
		{
			name:          "missing body",
			body:          "",
			boundary:      "",
			expectedResp:  true,
			expectedError: "request body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType := ""
			if tt.boundary != "" {
				contentType = "multipart/form-data; boundary=" + tt.boundary
			}

			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if contentType != "" {
						h.Set("Content-Type", contentType)
					}
				})

			// Only set body if there's actual content
			if tt.body != "" {
				req = req.Body(io.NopCloser(strings.NewReader(tt.body)))
			}

			request := req.Request()

			param := MultipartFormParam()
			result, resp := param.GetParamValue(context.Background(), request)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}

			if !tt.expectedResp {
				if result == nil {
					t.Error("expected multipart reader, got nil")
				}
				// Verify we can read from the multipart reader
				if result != nil {
					_, err := result.NextPart()
					if err != nil && err != io.EOF {
						t.Errorf("unexpected error reading multipart: %v", err)
					}
				}
			}
		})
	}
}

func TestJsonParam(t *testing.T) {
	type TestUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	tests := []struct {
		name          string
		body          string
		contentType   string
		expectedData  TestUser
		expectedResp  bool
		expectedError string
	}{
		{
			name:        "valid JSON",
			body:        `{"name":"John Doe","email":"john@example.com","age":30}`,
			contentType: "application/json",
			expectedData: TestUser{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			expectedResp: false,
		},
		{
			name:        "JSON with extra fields",
			body:        `{"name":"Jane","email":"jane@example.com","age":25,"extra":"ignored"}`,
			contentType: "application/json",
			expectedData: TestUser{
				Name:  "Jane",
				Email: "jane@example.com",
				Age:   25,
			},
			expectedResp: false,
		},
		{
			name:        "JSON with missing fields",
			body:        `{"name":"Bob"}`,
			contentType: "application/json",
			expectedData: TestUser{
				Name:  "Bob",
				Email: "",
				Age:   0,
			},
			expectedResp: false,
		},
		{
			name:          "invalid JSON",
			body:          `{"name":"Invalid JSON`,
			contentType:   "application/json",
			expectedData:  TestUser{},
			expectedResp:  true,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "missing body",
			body:          "",
			contentType:   "",
			expectedData:  TestUser{},
			expectedResp:  true,
			expectedError: "request body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest().
				ConfHeader(func(h *specs.Header) {
					if tt.contentType != "" {
						h.Set("Content-Type", tt.contentType)
					}
				})

			// Only set body if there's actual content
			if tt.body != "" {
				req = req.Body(io.NopCloser(strings.NewReader(tt.body)))
			}

			request := req.Request()

			param := JsonParam[TestUser]()
			result, resp := param.GetParamValue(context.Background(), request)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}

			if !tt.expectedResp {
				if !reflect.DeepEqual(result, tt.expectedData) {
					t.Errorf("expected %+v, got %+v", tt.expectedData, result)
				}
			}
		})
	}
}

func TestRawBodyParam(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedData  []byte
		expectedResp  bool
		expectedError string
	}{
		{
			name:         "valid body",
			body:         "Hello, World!",
			expectedData: []byte("Hello, World!"),
			expectedResp: false,
		},
		{
			name:          "empty body",
			body:          "",
			expectedData:  nil,
			expectedResp:  true,
			expectedError: "request body is required",
		},
		{
			name:         "binary data",
			body:         "\x00\x01\x02\x03\x04\x05",
			expectedData: []byte{0, 1, 2, 3, 4, 5},
			expectedResp: false,
		},
		{
			name:          "missing body",
			body:          "",
			expectedData:  nil,
			expectedResp:  true,
			expectedError: "request body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest()

			// Only set body if there's actual content
			if tt.body != "" {
				req = req.Body(io.NopCloser(strings.NewReader(tt.body)))
			}

			request := req.Request()

			param := RawBodyParam()
			result, resp := param.GetParamValue(context.Background(), request)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}

			if !tt.expectedResp {
				if !bytes.Equal(result, tt.expectedData) {
					t.Errorf("expected %v, got %v", tt.expectedData, result)
				}
			}
		})
	}
}

func TestStreamBodyParam(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		expectedResp  bool
		expectedError string
	}{
		{
			name:         "valid body",
			body:         "Streaming content",
			expectedResp: false,
		},
		{
			name:          "empty body",
			body:          "",
			expectedResp:  true,
			expectedError: "request body is required",
		},
		{
			name:         "large body",
			body:         strings.Repeat("Large content ", 1000),
			expectedResp: false,
		},
		{
			name:          "missing body",
			body:          "",
			expectedResp:  true,
			expectedError: "request body is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mock.DefaultRequest()

			// Only set body if there's actual content
			if tt.body != "" {
				req = req.Body(io.NopCloser(strings.NewReader(tt.body)))
			}

			request := req.Request()

			param := StreamBodyParam()
			result, resp := param.GetParamValue(context.Background(), request)

			if tt.expectedResp && resp == nil {
				t.Error("expected error response, got nil")
			}
			if !tt.expectedResp && resp != nil {
				t.Errorf("unexpected error response: %v", resp)
			}

			if !tt.expectedResp {
				if result == nil {
					t.Error("expected reader, got nil")
				}
				// Verify we can read from the stream
				if result != nil {
					content, err := io.ReadAll(result)
					if err != nil {
						t.Errorf("unexpected error reading stream: %v", err)
					}
					if string(content) != tt.body {
						t.Errorf("expected content %q, got %q", tt.body, string(content))
					}
				}
			}
		})
	}
}

func TestBodyParamEdgeCases(t *testing.T) {
	t.Run("empty content type", func(t *testing.T) {
		body := strings.NewReader("test content")
		req := mock.DefaultRequest().
			Body(io.NopCloser(body)).
			Request()

		// Test that body params work without content type
		param := RawBodyParam()
		result, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if result == nil {
			t.Error("expected body content, got nil")
		}
	})
}

func TestBodyParamWithMockRequest(t *testing.T) {
	t.Run("mock request with body", func(t *testing.T) {
		bodyContent := "test body content"
		req := mock.DefaultRequest().
			Body(io.NopCloser(strings.NewReader(bodyContent))).
			Request()

		// Test RawBodyParam with mock request
		param := RawBodyParam()
		result, resp := param.GetParamValue(context.Background(), req)

		if resp != nil {
			t.Errorf("unexpected error response: %v", resp)
		}
		if string(result) != bodyContent {
			t.Errorf("expected %q, got %q", bodyContent, string(result))
		}
	})
}
