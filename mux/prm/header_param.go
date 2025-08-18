package prm

import (
	"context"

	"github.com/oesand/plow"
)

// HeaderParam creates a new HeaderParameter for extracting and validating HTTP header values.
// It takes a header name and optional validation conditions.
func HeaderParam(name string, conditions ...Condition[string]) OptionalParameterProvider[string] {
	return &headerParameter{
		name:       name,
		conditions: conditions,
	}
}

type headerParameter struct {
	name       string
	required   bool
	conditions []Condition[string]
}

func (hp *headerParameter) Require() OptionalParameterProvider[string] {
	hp.required = true
	return hp
}

func (hp *headerParameter) GetParamValue(_ context.Context, req plow.Request) (string, plow.Response) {
	value := req.Header().Get(hp.name)

	var resp plow.Response
	if value == "" {
		if hp.required {
			resp = ErrorResponse("header '%s' is required", hp.name)
		}
		return value, resp
	}

	for _, condition := range hp.conditions {
		if err := condition.Validate(value); err != nil {
			resp = ErrorResponse("header '%s' is invalid: %s", hp.name, err)
			break
		}
	}

	return value, resp
}
