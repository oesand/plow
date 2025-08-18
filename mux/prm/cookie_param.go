package prm

import (
	"context"

	"github.com/oesand/plow"
)

// CookieParam creates a new cookieParameter for extracting and validating HTTP cookie values.
// It takes a cookie name and optional validation conditions.
func CookieParam(name string, conditions ...Condition[string]) OptionalParameterProvider[string] {
	return &cookieParameter{
		name:       name,
		conditions: conditions,
	}
}

type cookieParameter struct {
	name       string
	required   bool
	conditions []Condition[string]
}

func (cp *cookieParameter) Require() OptionalParameterProvider[string] {
	cp.required = true
	return cp
}

func (cp *cookieParameter) GetParamValue(_ context.Context, req plow.Request) (string, plow.Response) {
	var value string
	if cookie := req.Header().GetCookie(cp.name); cookie != nil {
		value = cookie.Value
	}

	var resp plow.Response
	if value == "" {
		if cp.required {
			resp = ErrorResponse("cookie '%s' is required", cp.name)
		}
		return value, resp
	}

	for _, condition := range cp.conditions {
		if err := condition.Validate(value); err != nil {
			resp = ErrorResponse("cookie '%s' is invalid: %s", cp.name, err)
			break
		}
	}

	return value, resp
}
