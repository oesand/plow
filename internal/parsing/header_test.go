package parsing

import (
	"bufio"
	"context"
	"github.com/oesand/giglet/specs"
	"reflect"
	"strings"
	"testing"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    *specs.Header
		wantErr bool
	}{
		{
			name: "Valid headers",
			text: "Content-Type: application/json\nAuthorization: Bearer token\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("Content-Type", "application/json")
				h.Set("Authorization", "Bearer token")
			}),
		},
		{
			name: "Empty input",
			text: "",
			want: specs.NewHeader(func(h *specs.Header) {}),
		},
		{
			name:    "Only whitespace lines",
			text:    "  \n\t\n",
			wantErr: true,
		},
		{
			name: "Missing colon",
			text: "Content-Type application/json\n",
			want: specs.NewHeader(),
		},
		{
			name: "Invalid character in key",
			text: "Bad@Key: value\n",
			want: specs.NewHeader(),
		},
		{
			name: "Header with trailing space",
			text: "Server: nginx   \n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("Server", "nginx")
			}),
		},
		{
			name: "Duplicate keys",
			text: "X-Test: 1\nX-Test: 2\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Test", "2") // assuming Set overrides
			}),
		},
		{
			name: "Value with colon",
			text: "X-Info: key:val\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Info", "key:val")
			}),
		},
		{
			name: "Header with underscore",
			text: "X_Custom_Header: ok\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X_Custom_Header", "ok")
			}),
		},
		{
			name: "Colon in key",
			text: "X:col: value\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X", "col: value")
			}),
		},
		{
			name: "Header with multiple spaces",
			text: "X-Name:   value with    spaces   \n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Name", "value with    spaces")
			}),
		},
		{
			name: "Mixed case key",
			text: "x-content-type: text/html\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Content-Type", "text/html")
			}),
		},
		{
			name: "Multiline header value with space continuation",
			text: "X-Note: This is line one\n and this is line two\n and finally line three\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Note", "This is line one and this is line two and finally line three")
			}),
		},
		{
			name: "Multiline header value with tab continuation",
			text: "X-Description: Start of value\n\tfollowed by tabbed line\n\tand another\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Description", "Start of value followed by tabbed line and another")
			}),
		},
		{
			name: "Multiline with mixed space and tab",
			text: "Folded-Header: First\n\tSecond line\n Third line\n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("Folded-Header", "First Second line Third line")
			}),
		},
		{
			name: "Multiline header with extra spaces preserved in value",
			text: "X-Long: first   \n  second   line  \n",
			want: specs.NewHeader(func(h *specs.Header) {
				h.Set("X-Long", "first second   line")
			}),
		},
		{
			name: "Continuation line without leading whitespace (invalid)",
			text: "X-Test: value\nnot-indented\n",
			want: specs.NewHeader(), // no leading space = should be parsed as a new header
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.text))
			got, err := ParseHeaders(ctx, reader, 0, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHeaders() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseHeaders() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ParseHeaderKVLine(t *testing.T) {
	tests := []struct {
		name      string
		line      []byte
		wantKey   []byte
		wantValue []byte
		wantFail  bool
	}{
		{
			name:      "Valid header with lowercase key",
			line:      []byte("content-type: text/html"),
			wantKey:   []byte("Content-Type"),
			wantValue: []byte("text/html"),
		},
		{
			name:      "Key with mixed separators",
			line:      []byte("x_custom-header: value123"),
			wantKey:   []byte("X_Custom-Header"),
			wantValue: []byte("value123"),
		},
		{
			name:     "Key with only colon",
			line:     []byte(": value"),
			wantFail: true,
		},
		{
			name:     "Missing colon (no separator)",
			line:     []byte("InvalidHeader"),
			wantFail: true,
		},
		{
			name:     "Empty line",
			line:     []byte(""),
			wantFail: true,
		},
		{
			name:     "Key with invalid char",
			line:     []byte("Bad@Key: value"),
			wantFail: true,
		},
		{
			name:      "Value with trailing space",
			line:      []byte("Server: nginx   "),
			wantKey:   []byte("Server"),
			wantValue: []byte("nginx"),
		},
		{
			name:      "Location header",
			line:      []byte("Location: http://example.org/test?key=value#fragment"),
			wantKey:   []byte("Location"),
			wantValue: []byte("http://example.org/test?key=value#fragment"),
		},
		{
			name:      "Value with tab padding",
			line:      []byte("X: \t  val\t\t "),
			wantKey:   []byte("X"),
			wantValue: []byte("val"),
		},
		{
			name:      "Value with colon",
			line:      []byte("X-Test: val:123"),
			wantKey:   []byte("X-Test"),
			wantValue: []byte("val:123"),
		},
		{
			name:      "Only key and colon",
			line:      []byte("X-Empty:"),
			wantKey:   []byte("X-Empty"),
			wantValue: []byte(""),
		},
		{
			name:      "Value with internal space",
			line:      []byte("X-Desc: some value"),
			wantKey:   []byte("X-Desc"),
			wantValue: []byte("some value"),
		},
		{
			name:      "Leading whitespace means all value",
			line:      []byte(" value"),
			wantValue: []byte("value"),
		},
		{
			name:      "Leading tab means all value",
			line:      []byte("\tvalue"),
			wantValue: []byte("value"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := parseHeaderKVLine(tt.line)
			if tt.wantFail == ok {
				t.Errorf("parseHeaderKVLine() got ok %v on line = %s, want ok %v", ok, tt.line, !tt.wantFail)
			}
			if !reflect.DeepEqual(key, tt.wantKey) {
				t.Errorf("parseHeaderKVLine() got key = %s, want %s", key, tt.wantKey)
			}
			if !reflect.DeepEqual(value, tt.wantValue) {
				t.Errorf("parseHeaderKVLine() got value = %s, want %s", value, tt.wantValue)
			}
		})
	}
}
