package specs

import (
	"testing"
)

func TestMatchContentType(t *testing.T) {
	tests := []struct {
		name     string
		header   *Header
		expected string
		wantOk   bool
	}{
		// Success cases
		{
			name: "Exact Match",
			header: NewHeader(func(header *Header) {
				header.Set("Content-Type", "application/json")
			}),
			expected: ContentTypeJson,
			wantOk:   true,
		},
		{
			name: "With Charset Parameter",
			header: NewHeader(func(header *Header) {
				header.Set("Content-Type", "application/json; charset=utf-8")
			}),
			expected: ContentTypeJson,
			wantOk:   true,
		},
		{
			name: "Case Insensitive",
			header: NewHeader(func(header *Header) {
				header.Set("Content-Type", "Application/JSON; Charset=UTF-8")
			}),
			expected: ContentTypeJson,
			wantOk:   true,
		},
		{
			name: "Extra Spaces",
			header: NewHeader(func(header *Header) {
				header.Set("Content-Type", "   application/json   ; charset=utf-8")
			}),
			expected: ContentTypeJson,
			wantOk:   true,
		},

		// Failed cases
		{
			name:     "Missing Header",
			header:   NewHeader(),
			expected: ContentTypeJson,
			wantOk:   false,
		},
		{
			name: "Other Type",
			header: NewHeader(func(header *Header) {
				header.Set("Content-Type", "text/html")
			}),
			expected: ContentTypeJson,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchContentType(tt.header, tt.expected)
			if got != tt.wantOk {
				t.Errorf("MatchContentType() = %v, want %v", got, tt.wantOk)
			}
		})
	}
}
