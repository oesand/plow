package plain

import (
	"reflect"
	"testing"
)

func Test_TitleCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"simple lowercase", "hello world", "Hello World"},
		{"mixed case", "gO is coOL", "Go Is Cool"},
		{"already title", "Go Is Great", "Go Is Great"},
		{"with punctuation", "hi, bob!", "Hi, Bob!"},
		{"multiple spaces", "a  b   c", "A  B   C"},
		{"newline separated", "line\nbreak", "Line\nBreak"},
		{"tabs and spaces", "foo\tbar baz", "Foo\tBar Baz"},
		{"non-alpha start", "123abc test", "123abc Test"},

		// HTTP header and cookie name cases
		{"http header style", "content-type", "Content-Type"},
		{"multiple hyphens", "x-custom-header-name", "X-Custom-Header-Name"},
		{"http header with spaces", "x forwarded for", "X Forwarded For"},
		{"cookie name with underscore", "session_id token_name", "Session_Id Token_Name"},
		{"mixed http header", "ACCEPT-language", "Accept-Language"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleCaseBytes([]byte(tt.input)); !reflect.DeepEqual(got, []byte(tt.want)) {
				t.Errorf("toTitleCaseBytes() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
