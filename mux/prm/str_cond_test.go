package prm

import (
	"errors"
	"fmt"
	"regexp"
	"testing"
)

func TestRegexPatternCondition(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    string
		expected error
	}{
		{"matches pattern", `^[a-z]+$`, "hello", nil},
		{"doesn't match pattern", `^[a-z]+$`, "Hello123", errors.New("value mismatch pattern")},
		{"empty string with pattern", `^[a-z]*$`, "", nil},
		{"email pattern valid", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "test@example.com", nil},
		{"email pattern invalid", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "invalid-email", errors.New("value mismatch pattern")},
		{"number pattern valid", `^\d+$`, "12345", nil},
		{"number pattern invalid", `^\d+$`, "123abc", errors.New("value mismatch pattern")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := RegexPattern(tt.pattern)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestRegexCondition(t *testing.T) {
	tests := []struct {
		name     string
		regex    *regexp.Regexp
		value    string
		expected error
	}{
		{"matches regex", regexp.MustCompile(`^[A-Z]+$`), "HELLO", nil},
		{"doesn't match regex", regexp.MustCompile(`^[A-Z]+$`), "Hello", errors.New("value mismatch pattern")},
		{"empty string with regex", regexp.MustCompile(`^.*$`), "", nil},
		{"complex regex valid", regexp.MustCompile(`^[a-z]{3,10}$`), "hello", nil},
		{"complex regex invalid", regexp.MustCompile(`^[a-z]{3,10}$`), "hi", errors.New("value mismatch pattern")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := Regex(tt.regex)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestLenCondition(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		value    string
		expected error
	}{
		{"exact length", 5, "hello", nil},
		{"too short", 5, "hi", fmt.Errorf("value must have exactly 5 characters")},
		{"too long", 5, "hello world", fmt.Errorf("value must have exactly 5 characters")},
		{"zero length", 0, "", nil},
		{"zero length with content", 0, "x", fmt.Errorf("value must have exactly 0 characters")},
		{"unicode characters", 3, "🚀🌍💻", nil},
		{"unicode characters wrong length", 3, "🚀🌍", fmt.Errorf("value must have exactly 3 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := Len(tt.length)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMinLenCondition(t *testing.T) {
	tests := []struct {
		name     string
		minLen   int
		value    string
		expected error
	}{
		{"exact minimum length", 5, "hello", nil},
		{"longer than minimum", 5, "hello world", nil},
		{"shorter than minimum", 5, "hi", fmt.Errorf("value must have at least 5 characters")},
		{"zero minimum length", 0, "", nil},
		{"zero minimum length with content", 0, "x", nil},
		{"unicode characters exact", 3, "🚀🌍💻", nil},
		{"unicode characters longer", 3, "🚀🌍💻🌟", nil},
		{"unicode characters shorter", 3, "🚀🌍", fmt.Errorf("value must have at least 3 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := MinLen(tt.minLen)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestMaxLenCondition(t *testing.T) {
	tests := []struct {
		name     string
		maxLen   int
		value    string
		expected error
	}{
		{"exact maximum length", 5, "hello", nil},
		{"shorter than maximum", 5, "hi", nil},
		{"longer than maximum", 5, "hello world", fmt.Errorf("value must have at most 5 characters")},
		{"zero maximum length", 0, "", nil},
		{"zero maximum length with content", 0, "x", fmt.Errorf("value must have at most 0 characters")},
		{"unicode characters exact", 3, "🚀🌍💻", nil},
		{"unicode characters shorter", 3, "🚀🌍", nil},
		{"unicode characters longer", 3, "🚀🌍💻🌟", fmt.Errorf("value must have at most 3 characters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := MaxLen(tt.maxLen)
			result := condition.Validate(tt.value)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected no error, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected error %v, got nil", tt.expected)
				} else if result.Error() != tt.expected.Error() {
					t.Errorf("expected error %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestStringConditionsEdgeCases(t *testing.T) {
	t.Run("regex with special characters", func(t *testing.T) {
		pattern := `^[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]+$`
		condition := RegexPattern(pattern)

		if err := condition.Validate("!@#$%"); err != nil {
			t.Error("expected no error for valid special characters")
		}
		if err := condition.Validate("abc123"); err == nil {
			t.Error("expected error for invalid characters")
		}
	})

	t.Run("length with mixed unicode", func(t *testing.T) {
		condition := Len(6)
		mixedString := "🚀🌍💻abc"

		if err := condition.Validate(mixedString); err != nil {
			t.Error("expected no error for mixed unicode string with correct length")
		}
	})

	t.Run("min length with empty string", func(t *testing.T) {
		condition := MinLen(1)
		if err := condition.Validate(""); err == nil {
			t.Error("expected error for empty string with min length 1")
		}
	})

	t.Run("max length with empty string", func(t *testing.T) {
		condition := MaxLen(5)
		if err := condition.Validate(""); err != nil {
			t.Error("expected no error for empty string with max length 5")
		}
	})

	t.Run("very long strings", func(t *testing.T) {
		longString := string(make([]rune, 1000))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}

		condition := MaxLen(1000)
		if err := condition.Validate(longString); err != nil {
			t.Error("expected no error for string at max length")
		}

		condition2 := MaxLen(999)
		if err := condition2.Validate(longString); err == nil {
			t.Error("expected error for string exceeding max length")
		}
	})
}

func TestStringConditionsCombined(t *testing.T) {
	t.Run("regex and length combination", func(t *testing.T) {
		// Test that regex condition works with length conditions
		regexCond := RegexPattern(`^[a-z]+$`)
		lenCond := Len(3)

		validString := "abc"
		if err := regexCond.Validate(validString); err != nil {
			t.Error("regex condition failed for valid string")
		}
		if err := lenCond.Validate(validString); err != nil {
			t.Error("length condition failed for valid string")
		}

		invalidString := "ab"
		if err := lenCond.Validate(invalidString); err == nil {
			t.Error("length condition should fail for invalid string")
		}
	})

	t.Run("min and max length boundary", func(t *testing.T) {
		minCond := MinLen(5)
		maxCond := MaxLen(10)

		boundaryString := "hello"
		if err := minCond.Validate(boundaryString); err != nil {
			t.Error("min length condition failed at boundary")
		}
		if err := maxCond.Validate(boundaryString); err != nil {
			t.Error("max length condition failed at boundary")
		}
	})
}
