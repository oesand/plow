package parsing

import "testing"

func Test_parseHTTPVersion(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantMajor uint16
		wantMinor uint16
		wantOk    bool
	}{
		{
			name:      "Valid HTTP/1.1",
			value:     "HTTP/1.1",
			wantMajor: 1,
			wantMinor: 1,
			wantOk:    true,
		},
		{
			name:      "Valid HTTP/2.0",
			value:     "HTTP/2.0",
			wantMajor: 2,
			wantMinor: 0,
			wantOk:    true,
		},
		{
			name:      "Valid HTTP/1.0",
			value:     "HTTP/1.0",
			wantMajor: 1,
			wantMinor: 0,
			wantOk:    true,
		},
		{
			name:      "Lowercase http",
			value:     "http/1.1",
			wantMajor: 1,
			wantMinor: 1,
			wantOk:    true,
		},
		{
			name:      "Extra whitespace",
			value:     "  HTTP/1.1  ",
			wantMajor: 1,
			wantMinor: 1,
			wantOk:    false,
		},
		{
			name:      "Missing slash",
			value:     "HTTP11",
			wantMajor: 0,
			wantMinor: 0,
			wantOk:    false,
		},
		{
			name:      "Only major version",
			value:     "HTTP/1",
			wantMajor: 0,
			wantMinor: 0,
			wantOk:    false,
		},
		{
			name:      "Non-numeric version",
			value:     "HTTP/a.b",
			wantMajor: 0,
			wantMinor: 0,
			wantOk:    false,
		},
		{
			name:      "Garbage input",
			value:     "banana",
			wantMajor: 0,
			wantMinor: 0,
			wantOk:    false,
		},
		{
			name:      "Empty input",
			value:     "",
			wantMajor: 0,
			wantMinor: 0,
			wantOk:    false,
		},
		{
			name:      "Valid large version",
			value:     "HTTP/10.42",
			wantMajor: 10,
			wantMinor: 42,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor, gotOk := parseHTTPVersion([]byte(tt.value))
			if gotOk != tt.wantOk {
				t.Errorf("parseHTTPVersion() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if tt.wantOk {
				if gotMajor != tt.wantMajor {
					t.Errorf("parseHTTPVersion() gotMajor = %v, want %v", gotMajor, tt.wantMajor)
				}
				if gotMinor != tt.wantMinor {
					t.Errorf("parseHTTPVersion() gotMinor = %v, want %v", gotMinor, tt.wantMinor)
				}
			}
		})
	}
}
