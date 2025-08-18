package prm

import (
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"
)

// RegexPattern creates a condition that validates a string matches the given regular expression pattern.
func RegexPattern(pattern string) Condition[string] {
	regex := regexp.MustCompile(pattern)
	return &regexCond{regex}
}

// Regex creates a condition that validates a string matches the given regular expression.
func Regex(regex *regexp.Regexp) Condition[string] {
	if regex == nil {
		panic("regex is nil")
	}
	return &regexCond{regex}
}

type regexCond struct {
	regex *regexp.Regexp
}

func (c *regexCond) Validate(value string) error {
	if !c.regex.MatchString(value) {
		return errors.New("value mismatch pattern")
	}
	return nil
}

// Len creates a condition that validates a string has exactly the specified length.
func Len(length int) Condition[string] {
	return &lenCond{length}
}

type lenCond struct {
	length int
}

func (c *lenCond) Validate(value string) error {
	if utf8.RuneCountInString(value) != c.length {
		return fmt.Errorf("value must have exactly %d characters", c.length)
	}
	return nil
}

// MinLen creates a condition that validates a string has at least the specified length.
func MinLen(minLength int) Condition[string] {
	return &minLenCond{minLength}
}

type minLenCond struct {
	minLength int
}

func (c *minLenCond) Validate(value string) error {
	if utf8.RuneCountInString(value) < c.minLength {
		return fmt.Errorf("value must have at least %d characters", c.minLength)
	}
	return nil
}

// MaxLen creates a condition that validates a string has at most the specified length.
func MaxLen(maxLength int) Condition[string] {
	return &maxLenCond{maxLength}
}

type maxLenCond struct {
	maxLength int
}

func (c *maxLenCond) Validate(value string) error {
	if utf8.RuneCountInString(value) > c.maxLength {
		return fmt.Errorf("value must have at most %d characters", c.maxLength)
	}
	return nil
}
