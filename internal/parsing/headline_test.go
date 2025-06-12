package parsing

import (
	"github.com/oesand/giglet/specs"
	"testing"
)

func TestParseClientRequestHeadline(t *testing.T) {
	tests := []struct {
		name       string
		headline   string
		wantMethod specs.HttpMethod
		wantUrl    string
		wantMajor  uint16
		wantMinor  uint16
		wantOk     bool
	}{
		{
			name:       "Valid GET request",
			headline:   "GET /index.html HTTP/1.1",
			wantMethod: specs.HttpMethodGet,
			wantUrl:    "/index.html",
			wantMajor:  1,
			wantMinor:  1,
			wantOk:     true,
		},
		{
			name:       "Valid POST request",
			headline:   "POST /submit HTTP/1.0",
			wantMethod: specs.HttpMethodPost,
			wantUrl:    "/submit",
			wantMajor:  1,
			wantMinor:  0,
			wantOk:     true,
		},
		{
			name:       "Valid PUT request with query string",
			headline:   "PUT /update?id=42 HTTP/2.0",
			wantMethod: specs.HttpMethodPut,
			wantUrl:    "/update?id=42",
			wantMajor:  2,
			wantMinor:  0,
			wantOk:     true,
		},
		{
			name:       "Lowercase method (invalid)",
			headline:   "get / HTTP/1.1",
			wantMethod: "get",
			wantUrl:    "/",
			wantMajor:  1,
			wantMinor:  1,
			wantOk:     true,
		},
		{
			name:     "Missing HTTP version",
			headline: "GET /index.html",
			wantOk:   false,
		},
		{
			name:     "Malformed version",
			headline: "GET /index.html HTTP/one.one",
			wantOk:   false,
		},
		{
			name:     "Missing URL",
			headline: "GET  HTTP/1.1",
			wantOk:   false,
		},
		{
			name:     "Extra spaces",
			headline: "  GET   /home   HTTP/1.1  ",
			wantOk:   false,
		},
		{
			name:       "Unknown method",
			headline:   "FOO /bar HTTP/1.1",
			wantMethod: "FOO",
			wantUrl:    "/bar",
			wantMajor:  1,
			wantMinor:  1,
			wantOk:     true,
		},
		{
			name:     "Empty headline",
			headline: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMethod, gotUrl, gotMajor, gotMinor, gotOk := ParseClientRequestHeadline([]byte(tt.headline))

			if gotOk != tt.wantOk {
				t.Errorf("ParseClientRequestHeadline() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk {
				if gotMethod != tt.wantMethod {
					t.Errorf("ParseClientRequestHeadline() gotMethod = %v, want %v", gotMethod, tt.wantMethod)
				}
				if gotUrl != tt.wantUrl {
					t.Errorf("ParseClientRequestHeadline() gotUrl = %v, want %v", gotUrl, tt.wantUrl)
				}
				if gotMajor != tt.wantMajor {
					t.Errorf("ParseClientRequestHeadline() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
				}
				if gotMinor != tt.wantMinor {
					t.Errorf("ParseClientRequestHeadline() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
				}
			}
		})
	}
}

func TestParseServerResponseHeadline(t *testing.T) {
	tests := []struct {
		name       string
		headline   string
		wantStatus specs.StatusCode
		wantMajor  uint16
		wantMinor  uint16
		wantRes    bool
	}{
		{
			name:       "Standard OK response",
			headline:   "HTTP/1.1 200 OK",
			wantStatus: specs.StatusCodeOK,
			wantMajor:  1,
			wantMinor:  1,
			wantRes:    true,
		},
		{
			name:       "HTTP/2 404 Not Found",
			headline:   "HTTP/2.0 404 Not Found",
			wantStatus: specs.StatusCodeNotFound,
			wantMajor:  2,
			wantMinor:  0,
			wantRes:    true,
		},
		{
			name:       "HTTP/1.0 500 Internal Server Error",
			headline:   "HTTP/1.0 500 Internal Server Error",
			wantStatus: specs.StatusCodeInternalServerError,
			wantMajor:  1,
			wantMinor:  0,
			wantRes:    true,
		},
		{
			name:     "Extra spaces",
			headline: "  HTTP/1.1    204    No Content  ",
			wantRes:  false,
		},
		{
			name:       "Invalid version format",
			headline:   "HTTP/one.one 200 OK",
			wantStatus: 0,
			wantMajor:  0,
			wantMinor:  0,
			wantRes:    false,
		},
		{
			name:       "Missing status code",
			headline:   "HTTP/1.1 OK",
			wantStatus: 0,
			wantMajor:  0,
			wantMinor:  0,
			wantRes:    false,
		},
		{
			name:       "Garbage input",
			headline:   "banana sandwich",
			wantStatus: 0,
			wantMajor:  0,
			wantMinor:  0,
			wantRes:    false,
		},
		{
			name:       "Empty headline",
			headline:   "",
			wantStatus: 0,
			wantMajor:  0,
			wantMinor:  0,
			wantRes:    false,
		},
		{
			name:       "Unknown status code",
			headline:   "HTTP/1.1 999 Custom Status",
			wantStatus: specs.StatusCode(999),
			wantMajor:  1,
			wantMinor:  1,
			wantRes:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotMajor, gotMinor, gotRes := ParseServerResponseHeadline([]byte(tt.headline))
			if gotRes != tt.wantRes {
				t.Errorf("ParseServerResponseHeadline() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
			if tt.wantRes {
				if gotStatus != tt.wantStatus {
					t.Errorf("ParseServerResponseHeadline() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
				}
				if gotMajor != tt.wantMajor {
					t.Errorf("ParseServerResponseHeadline() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
				}
				if gotMinor != tt.wantMinor {
					t.Errorf("ParseServerResponseHeadline() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
				}
			}
		})
	}
}
