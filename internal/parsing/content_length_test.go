package parsing

import (
	"github.com/oesand/plow/specs"
	"testing"
)

func TestParseContentLength(t *testing.T) {
	tests := []struct {
		name          string
		header        *specs.Header
		wantIsChunked bool
		wantSize      int64
		wantErr       bool
	}{
		// Valid cases
		{
			name: "Only Transfer-Encoding",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Transfer-Encoding", "chunked")
			}),
			wantIsChunked: true,
		},
		{
			name: "Transfer-Encoding and Content-Length",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Transfer-Encoding", "chunked")
				header.Set("Content-Length", "100")
			}),
			wantIsChunked: true,
		},
		{
			name: "Only Content-Length",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Content-Length", "100")
			}),
			wantSize: 100,
		},

		// Invalid cases
		{
			name: "Invalid Transfer-Encoding",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Transfer-Encoding", "xxx")
			}),
			wantErr: true,
		},
		{
			name: "Invalid Transfer-Encoding and Valid Content-Length",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Transfer-Encoding", "xxx")
				header.Set("Content-Length", "100")
			}),
			wantErr: true,
		},
		{
			name: "Invalid Transfer-Encoding and Content-Length",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Transfer-Encoding", "xxx")
				header.Set("Content-Length", "xxx")
			}),
			wantErr: true,
		},
		{
			name: "Invalid Content-Length",
			header: specs.NewHeader(func(header *specs.Header) {
				header.Set("Content-Length", "xxx")
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsChunked, gotSize, err := ParseContentLength(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseContentLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIsChunked != tt.wantIsChunked {
				t.Errorf("ParseContentLength() gotIsChunked = %v, want %v", gotIsChunked, tt.wantIsChunked)
			}
			if gotSize != tt.wantSize {
				t.Errorf("ParseContentLength() gotSize = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}
