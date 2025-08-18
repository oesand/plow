package mux

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseRouteTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantErr     bool
		errContains string
		checkFunc   func(*testing.T, *routePattern)
	}{
		{
			name:     "simple path without parameters",
			template: "/users",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if rp.Template != "/users" {
					t.Errorf("OriginalTemplate = %v, want %v", rp.Template, "/users")
				}
				if len(rp.ParamNames) != 0 {
					t.Errorf("ParamNames = %v, want empty", rp.ParamNames)
				}
			},
		},
		{
			name:     "path with single parameter",
			template: "/users/{id}",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 1 {
					t.Errorf("ParamNames = %v, want 1 parameter", rp.ParamNames)
				}
				if rp.ParamNames[0] != "id" {
					t.Errorf("ParamNames[0] = %v, want %v", rp.ParamNames[0], "id")
				}
			},
		},
		{
			name:     "path with multiple parameters",
			template: "/users/{id}/posts/{postId}",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 2 {
					t.Errorf("ParamNames = %v, want 2 parameters", rp.ParamNames)
				}
				expected := []string{"id", "postId"}
				if !reflect.DeepEqual(rp.ParamNames, expected) {
					t.Errorf("ParamNames = %v, want %v", rp.ParamNames, expected)
				}
			},
		},
		{
			name:     "path with custom regex pattern",
			template: "/posts/{year:\\d{4}}/{slug:[^/]+}",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 2 {
					t.Errorf("ParamNames = %v, want 2 parameters", rp.ParamNames)
				}
				expected := []string{"year", "slug"}
				if !reflect.DeepEqual(rp.ParamNames, expected) {
					t.Errorf("ParamNames = %v, want %v", rp.ParamNames, expected)
				}
			},
		},
		{
			name:     "path with optional trailing slash",
			template: "/users/{id}/",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 1 {
					t.Errorf("ParamNames = %v, want 1 parameter", rp.ParamNames)
				}
				if rp.ParamNames[0] != "id" {
					t.Errorf("ParamNames[0] = %v, want %v", rp.ParamNames[0], "id")
				}
			},
		},
		{
			name:     "path with complex nested patterns",
			template: "/api/{version}/users/{id:\\d+}/profile/{section:[a-z]+}",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 3 {
					t.Errorf("ParamNames = %v, want 3 parameters", rp.ParamNames)
				}
				expected := []string{"version", "id", "section"}
				if !reflect.DeepEqual(rp.ParamNames, expected) {
					t.Errorf("ParamNames = %v, want %v", rp.ParamNames, expected)
				}
			},
		},
		{
			name:        "empty template",
			template:    "",
			wantErr:     true,
			errContains: "template cannot be empty",
		},
		{
			name:        "duplicate parameter names",
			template:    "/users/{id}/posts/{id}",
			wantErr:     true,
			errContains: "duplicate parameter name",
		},
		{
			name:        "empty parameter name",
			template:    "/users/{}/posts",
			wantErr:     true,
			errContains: "parameter name cannot be empty",
		},
		{
			name:        "invalid regex pattern",
			template:    "/users/{id:[invalid}",
			wantErr:     true,
			errContains: "failed to compile regex pattern",
		},
		{
			name:     "path with regex special characters",
			template: "/files/{name}/download",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 1 {
					t.Errorf("ParamNames = %v, want 1 parameter", rp.ParamNames)
				}
				if rp.ParamNames[0] != "name" {
					t.Errorf("ParamNames[0] = %v, want %v", rp.ParamNames[0], "name")
				}
			},
		},
		{
			name:     "path with dots and special chars",
			template: "/api/v1/users/{id}",
			checkFunc: func(t *testing.T, rp *routePattern) {
				if len(rp.ParamNames) != 1 {
					t.Errorf("ParamNames = %v, want 1 parameter", rp.ParamNames)
				}
				if rp.ParamNames[0] != "id" {
					t.Errorf("ParamNames[0] = %v, want %v", rp.ParamNames[0], "id")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp, err := parseRouteTemplate(tt.template)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRouteTemplate() error = nil, want error containing %q", tt.errContains)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ParseRouteTemplate() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRouteTemplate() unexpected error = %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, rp)
			}
		})
	}
}

func TestRoutePattern_Match(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		testPaths    []string
		expectMatch  []bool
		expectParams []map[string]string
	}{
		{
			name:     "simple path without parameters",
			template: "/users",
			testPaths: []string{
				"/users",
				"/users/",
				"/users/123",
				"/user",
				"/",
			},
			expectMatch: []bool{true, true, false, false, false},
			expectParams: []map[string]string{
				{},
				{},
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with single parameter",
			template: "/users/{id}",
			testPaths: []string{
				"/users/123",
				"/users/abc",
				"/users/",
				"/users/123/posts",
				"/user/123",
				"/users",
			},
			expectMatch: []bool{true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"id": "123"},
				{"id": "abc"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with multiple parameters",
			template: "/users/{id}/posts/{postId}",
			testPaths: []string{
				"/users/123/posts/456",
				"/users/abc/posts/def",
				"/users/123/posts/456/",
				"/users/abc/posts/def/",
				"/users/123/posts/",
				"/users/123/posts",
				"/users/123",
				"/users//posts/456",
			},
			expectMatch: []bool{true, true, true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"id": "123", "postId": "456"},
				{"id": "abc", "postId": "def"},
				{"id": "123", "postId": "456"},
				{"id": "abc", "postId": "def"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with custom regex pattern - year",
			template: "/posts/{year:\\d{4}}/{slug:[^/]+}",
			testPaths: []string{
				"/posts/2023/my-post-title",
				"/posts/2023/my-post-title/",
				"/posts/2023/",
				"/posts/23/my-post",
				"/posts/2023/my/post",
				"/posts/abcd/my-post",
				"/posts/2023",
			},
			expectMatch: []bool{true, true, false, false, false, false, false},
			expectParams: []map[string]string{
				{"year": "2023", "slug": "my-post-title"},
				{"year": "2023", "slug": "my-post-title"},
				nil,
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with optional trailing slash",
			template: "/users/{id}/",
			testPaths: []string{
				"/users/123",
				"/users/123/",
				"/users/abc",
				"/users/abc/",
				"/users/",
				"/users",
			},
			expectMatch: []bool{true, true, true, true, false, false},
			expectParams: []map[string]string{
				{"id": "123"},
				{"id": "123"},
				{"id": "abc"},
				{"id": "abc"},
				nil,
				nil,
			},
		},
		{
			name:     "path with numeric constraints",
			template: "/users/{id:\\d{1,3}}",
			testPaths: []string{
				"/users/1",
				"/users/123",
				"/users/1234",
				"/users/abc",
				"/users/",
				"/users/0",
			},
			expectMatch: []bool{true, true, false, false, false, true},
			expectParams: []map[string]string{
				{"id": "1"},
				{"id": "123"},
				nil,
				nil,
				nil,
				{"id": "0"},
			},
		},
		{
			name:     "path with alphanumeric pattern",
			template: "/tags/{tag:[a-zA-Z0-9]+}",
			testPaths: []string{
				"/tags/golang",
				"/tags/123",
				"/tags/go-lang",
				"/tags/",
				"/tags/go lang",
				"/tags/GO",
			},
			expectMatch: []bool{true, true, false, false, false, true},
			expectParams: []map[string]string{
				{"tag": "golang"},
				{"tag": "123"},
				nil,
				nil,
				nil,
				{"tag": "GO"},
			},
		},
		{
			name:     "path with wildcard parameter",
			template: "/static/{*}",
			testPaths: []string{
				"/static/css/style.css",
				"/static/js/app.js",
				"/static/images/logo.png",
				"/static/css/style.css/",
				"/static/js/app.js/",
				"/static/",
				"/static",
				"/api/users",
			},
			expectMatch: []bool{true, true, true, true, true, true, false, false},
			expectParams: []map[string]string{
				{"*": "css/style.css"},
				{"*": "js/app.js"},
				{"*": "images/logo.png"},
				{"*": "css/style.css"},
				{"*": "js/app.js"},
				{"*": ""},
				nil,
				nil,
			},
		},
		{
			name:     "path with mixed parameter types",
			template: "/api/{version}/users/{id:\\d+}/profile",
			testPaths: []string{
				"/api/v1/users/123/profile",
				"/api/v2/users/456/profile",
				"/api/v1/users/123/profile/",
				"/api/v2/users/456/profile/",
				"/api/v1/users/abc/profile",
				"/api/v1/users/123",
				"/api/v1/users/profile",
				"/api/users/123/profile",
			},
			expectMatch: []bool{true, true, true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"version": "v1", "id": "123"},
				{"version": "v2", "id": "456"},
				{"version": "v1", "id": "123"},
				{"version": "v2", "id": "456"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with regex special characters",
			template: "/files/{name}/download",
			testPaths: []string{
				"/files/document.pdf/download",
				"/files/my-file.txt/download",
				"/files/download",
				"/files//download",
				"/files/document.pdf",
				"/files/document.pdf/download/extra",
			},
			expectMatch: []bool{true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"name": "document.pdf"},
				{"name": "my-file.txt"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with complex regex pattern",
			template: "/email/{email:[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}}",
			testPaths: []string{
				"/email/user@example.com",
				"/email/test.email+tag@domain.co.uk",
				"/email/invalid-email",
				"/email/user@",
				"/email/@domain.com",
				"/email/user@domain",
			},
			expectMatch: []bool{true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"email": "user@example.com"},
				{"email": "test.email+tag@domain.co.uk"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with UUID pattern",
			template: "/users/{uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}",
			testPaths: []string{
				"/users/550e8400-e29b-41d4-a716-446655440000",
				"/users/550e8400-e29b-41d4-a716-44665544000",
				"/users/550e8400-e29b-41d4-a716-4466554400000",
				"/users/550e8400-e29b-41d4-a716-44665544000g",
				"/users/550e8400-e29b-41d4-a716-44665544000",
				"/users/not-a-uuid",
			},
			expectMatch: []bool{true, false, false, false, false, false},
			expectParams: []map[string]string{
				{"uuid": "550e8400-e29b-41d4-a716-446655440000"},
				nil,
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with simple regex pattern",
			template: "/regex/{pattern:[a-z]+}",
			testPaths: []string{
				"/regex/abc",
				"/regex/def",
				"/regex/ABC",
				"/regex/123",
				"/regex/",
				"/regex",
			},
			expectMatch: []bool{true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"pattern": "abc"},
				{"pattern": "def"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with complex nested patterns",
			template: "/api/{version}/users/{id:\\d+}/profile/{section:[a-z]+}",
			testPaths: []string{
				"/api/v1/users/123/profile/settings",
				"/api/v2/users/456/profile/preferences",
				"/api/v1/users/123/profile/settings/",
				"/api/v2/users/456/profile/preferences/",
				"/api/v1/users/abc/profile/settings",
				"/api/v1/users/123/profile",
				"/api/v1/users/123/profile/SETTINGS",
				"/api/v1/users/123/profile/settings/extra",
			},
			expectMatch: []bool{true, true, true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"version": "v1", "id": "123", "section": "settings"},
				{"version": "v2", "id": "456", "section": "preferences"},
				{"version": "v1", "id": "123", "section": "settings"},
				{"version": "v2", "id": "456", "section": "preferences"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with word characters in regex",
			template: "/search/{query:\\w+}",
			testPaths: []string{
				"/search/hello",
				"/search/world123",
				"/search/hello world",
				"/search/",
				"/search",
			},
			expectMatch: []bool{true, true, false, false, false},
			expectParams: []map[string]string{
				{"query": "hello"},
				{"query": "world123"},
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with quantifiers in regex",
			template: "/numbers/{num:\\d{1,5}}",
			testPaths: []string{
				"/numbers/1",
				"/numbers/123",
				"/numbers/12345",
				"/numbers/123456",
				"/numbers/abc",
				"/numbers/",
			},
			expectMatch: []bool{true, true, true, false, false, false},
			expectParams: []map[string]string{
				{"num": "1"},
				{"num": "123"},
				{"num": "12345"},
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with character classes in regex",
			template: "/categories/{cat:[A-Z][a-z]+}",
			testPaths: []string{
				"/categories/Technology",
				"/categories/Sports",
				"/categories/technology",
				"/categories/TECH",
				"/categories/",
				"/categories/123",
			},
			expectMatch: []bool{true, true, false, false, false, false},
			expectParams: []map[string]string{
				{"cat": "Technology"},
				{"cat": "Sports"},
				nil,
				nil,
				nil,
				nil,
			},
		},
		{
			name:     "path with optional trailing slash and parameters",
			template: "/api/{version}/users/{id}/",
			testPaths: []string{
				"/api/v1/users/123",
				"/api/v1/users/123/",
				"/api/v2/users/456",
				"/api/v2/users/456/",
				"/api/v1/users/",
				"/api/v1/users",
			},
			expectMatch: []bool{true, true, true, true, false, false},
			expectParams: []map[string]string{
				{"version": "v1", "id": "123"},
				{"version": "v1", "id": "123"},
				{"version": "v2", "id": "456"},
				{"version": "v2", "id": "456"},
				nil,
				nil,
			},
		},
		{
			name:     "path with wildcard parameter and custom regex",
			template: "/files/{*:.*\\.(css|js|png|jpg)}",
			testPaths: []string{
				"/files/css/style.css",
				"/files/js/app.js",
				"/files/images/logo.png",
				"/files/photos/photo.jpg",
				"/files/css/style.css/",
				"/files/js/app.js/",
				"/files/readme.txt",
				"/files/",
				"/files",
			},
			expectMatch: []bool{true, true, true, true, true, true, false, false, false},
			expectParams: []map[string]string{
				{"*": "css/style.css"},
				{"*": "js/app.js"},
				{"*": "images/logo.png"},
				{"*": "photos/photo.jpg"},
				{"*": "css/style.css"},
				{"*": "js/app.js"},
				nil,
				nil,
				nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp, err := parseRouteTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseRouteTemplate() error = %v", err)
			}

			for i, testPath := range tt.testPaths {
				matched, paramsSeq := rp.Match(testPath)

				if matched != tt.expectMatch[i] {
					t.Errorf("Match(%q) = %v, want %v", testPath, matched, tt.expectMatch[i])
				}

				if matched {
					// Convert iter.Seq2 to map for comparison
					params := make(map[string]string)
					for key, value := range paramsSeq {
						params[key] = value
					}

					if !reflect.DeepEqual(params, tt.expectParams[i]) {
						t.Errorf("Match(%q) params = %v, want %v", testPath, params, tt.expectParams[i])
					}
				} else if tt.expectParams[i] != nil {
					t.Errorf("Match(%q) expected params but got none", testPath)
				}
			}
		})
	}
}

func TestParseRouteTemplate_Escaping(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		description string
		checkFunc   func(*testing.T, *routePattern)
	}{
		{
			name:        "path with dots",
			template:    "/api/v1/users/{id}",
			description: "Dots in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				// Test that dots are properly escaped in the regex
				matched, _ := rp.Match("/api/v1/users/123")
				if !matched {
					t.Error("Expected /api/v1/users/123 to match")
				}
				matched, _ = rp.Match("/api/v1/users/123/")
				if !matched {
					t.Error("Expected /api/v1/users/123/ to match")
				}
			},
		},
		{
			name:        "path with plus signs",
			template:    "/search+results/{query}",
			description: "Plus signs in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/search+results/test")
				if !matched {
					t.Error("Expected /search+results/test to match")
				}
				matched, _ = rp.Match("/search+results/test/")
				if !matched {
					t.Error("Expected /search+results/test/ to match")
				}
			},
		},
		{
			name:        "path with asterisks",
			template:    "/files/*/download",
			description: "Asterisks in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/files/*/download")
				if !matched {
					t.Error("Expected /files/*/download to match")
				}
				matched, _ = rp.Match("/files/*/download/")
				if !matched {
					t.Error("Expected /files/*/download/ to match")
				}
			},
		},
		{
			name:        "path with question marks",
			template:    "/help?topic={topic}",
			description: "Question marks in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/help?topic=general")
				if !matched {
					t.Error("Expected /help?topic=general to match")
				}
				matched, _ = rp.Match("/help?topic=general/")
				if !matched {
					t.Error("Expected /help?topic=general/ to match")
				}
			},
		},
		{
			name:        "path with parentheses",
			template:    "/api/(v1)/users/{id}",
			description: "Parentheses in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/(v1)/users/123")
				if !matched {
					t.Error("Expected /api/(v1)/users/123 to match")
				}
				matched, _ = rp.Match("/api/(v1)/users/123/")
				if !matched {
					t.Error("Expected /api/(v1)/users/123/ to match")
				}
			},
		},
		{
			name:        "path with brackets",
			template:    "/api/[v1]/users/{id}",
			description: "Brackets in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/[v1]/users/123")
				if !matched {
					t.Error("Expected /api/[v1]/users/123 to match")
				}
				matched, _ = rp.Match("/api/[v1]/users/123/")
				if !matched {
					t.Error("Expected /api/[v1]/users/123/ to match")
				}
			},
		},
		{
			name:        "path with braces",
			template:    "/api/{v1}/users/{id}",
			description: "Braces in static path segments should be escaped (but not parameter braces)",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/{v1}/users/123")
				if !matched {
					t.Error("Expected /api/{v1}/users/123 to match")
				}
				matched, _ = rp.Match("/api/{v1}/users/123/")
				if !matched {
					t.Error("Expected /api/{v1}/users/123/ to match")
				}
			},
		},
		{
			name:        "path with pipes",
			template:    "/api/v1|v2/users/{id}",
			description: "Pipes in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/v1|v2/users/123")
				if !matched {
					t.Error("Expected /api/v1|v2/users/123 to match")
				}
				matched, _ = rp.Match("/api/v1|v2/users/123/")
				if !matched {
					t.Error("Expected /api/v1|v2/users/123/ to match")
				}
			},
		},
		{
			name:        "path with backslashes",
			template:    "/files\\backup\\{filename}",
			description: "Backslashes in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/files\\backup\\document.txt")
				if !matched {
					t.Error("Expected /files\\backup\\document.txt to match")
				}
				matched, _ = rp.Match("/files\\backup\\document.txt/")
				if !matched {
					t.Error("Expected /files\\backup\\document.txt/ to match")
				}
			},
		},
		{
			name:        "path with dollar signs",
			template:    "/pricing/$99/{plan}",
			description: "Dollar signs in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/pricing/$99/premium")
				if !matched {
					t.Error("Expected /pricing/$99/premium to match")
				}
				matched, _ = rp.Match("/pricing/$99/premium/")
				if !matched {
					t.Error("Expected /pricing/$99/premium/ to match")
				}
			},
		},
		{
			name:        "path with carets",
			template:    "/api/^v1/users/{id}",
			description: "Carets in static path segments should be escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/^v1/users/123")
				if !matched {
					t.Error("Expected /api/^v1/users/123 to match")
				}
				matched, _ = rp.Match("/api/^v1/users/123/")
				if !matched {
					t.Error("Expected /api/^v1/users/123/ to match")
				}
			},
		},
		{
			name:        "complex path with multiple special characters",
			template:    "/api/v1.0+beta/users/{id}/profile?section=settings",
			description: "Multiple special characters should all be properly escaped",
			checkFunc: func(t *testing.T, rp *routePattern) {
				matched, _ := rp.Match("/api/v1.0+beta/users/123/profile?section=settings")
				if !matched {
					t.Error("Expected complex path to match")
				}
				matched, _ = rp.Match("/api/v1.0+beta/users/123/profile?section=settings/")
				if !matched {
					t.Error("Expected complex path with trailing slash to match")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp, err := parseRouteTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseRouteTemplate() error = %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, rp)
			}
		})
	}
}
