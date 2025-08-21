package routing

import (
	"fmt"
	"iter"
	"regexp"
	"strings"
)

// RoutePattern represents a compiled route pattern with parameters
type RoutePattern struct {
	Original   string
	Regex      *regexp.Regexp
	ParamNames []string
	Depth      int
}

// ParseRoutePattern converts a route template into a regex pattern and parameter names
// Supported formats:
//   - /users/{id}
//   - /files/{name}/raw
//   - /posts/{year:\d{4}}/{slug:[^/]+}
//   - /api/{version}/users/{id:\d+}/profile
//   - /static/{*:.*} (wildcard parameter for "at the end")
//
// Everything outside of {…} is safely regex-escaped
// Trailing slash is ignored at compile-time; both /path and /path/ are accepted at match-time
// Wildcard parameters (*) can match any characters including slashes
func ParseRoutePattern(pattern string) (*RoutePattern, error) {
	if pattern == "" {
		return nil, fmt.Errorf("template cannot be empty")
	}

	// Normalize: ignore a single trailing slash in the template
	normalized := strings.TrimSuffix(pattern, "/")

	spans := findPlaceholders(normalized)

	var paramNames []string
	var b strings.Builder
	last := 0

	for _, span := range spans {
		start, end := span[0], span[1]
		if start > last {
			b.WriteString(regexp.QuoteMeta(normalized[last:start]))
		}
		content := normalized[start+1 : end-1]
		parts := strings.SplitN(content, ":", 2)
		name := strings.TrimSpace(parts[0])
		if name == "" {
			return nil, fmt.Errorf("parameter name cannot be empty")
		}
		for _, n := range paramNames {
			if n == name {
				return nil, fmt.Errorf("duplicate parameter name: %s", name)
			}
		}
		paramNames = append(paramNames, name)

		// Prefer explicit custom regex if provided
		var pattern string
		if len(parts) == 2 {
			pattern = parts[1]
		} else if name == "*" {
			// Wildcard parameter can match anything including slashes; use non-greedy to avoid swallowing optional trailing slash
			pattern = ".*?"
		} else {
			pattern = "[^/]+"
		}

		b.WriteByte('(')
		b.WriteString(pattern)
		b.WriteByte(')')
		last = end
	}
	if last < len(normalized) {
		b.WriteString(regexp.QuoteMeta(normalized[last:]))
	}

	regexPattern := b.String()
	if !strings.HasPrefix(regexPattern, "^") {
		regexPattern = "^" + regexPattern
	}
	// Always allow an optional trailing slash at match-time
	regexPattern += "/?"
	if !strings.HasSuffix(regexPattern, "$") {
		regexPattern += "$"
	}

	compiledRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex pattern: %w", err)
	}
	depth := strings.Count(pattern, "/")

	return &RoutePattern{
		Original:   normalized,
		Regex:      compiledRegex,
		ParamNames: paramNames,
		Depth:      depth,
	}, nil
}

// findPlaceholders returns ranges [start,end) for balanced {...} placeholders in the template
func findPlaceholders(s string) [][2]int {
	var res [][2]int
	depth := 0
	start := -1
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if depth == 0 {
				start = i
			}
			depth++
		case '}':
			if depth > 0 {
				depth--
				if depth == 0 && start >= 0 {
					res = append(res, [2]int{start, i + 1})
					start = -1
				}
			}
		}
	}
	return res
}

func (rp *RoutePattern) Match(path string) (bool, iter.Seq2[string, string]) {
	if rp.Regex == nil {
		return false, nil
	}

	matches := rp.Regex.FindStringSubmatch(path)
	if matches == nil {
		return false, nil
	}

	// Return parameter values as iter.Seq2
	return true, func(yield func(string, string) bool) {
		for i, paramName := range rp.ParamNames {
			if i+1 < len(matches) {
				value := matches[i+1]
				// Special-case wildcard parameter to avoid capturing trailing slash
				if paramName == "*" {
					value = strings.TrimSuffix(value, "/")
				}
				if !yield(paramName, value) {
					return
				}
			}
		}
	}
}
