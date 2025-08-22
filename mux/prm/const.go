package prm

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// NumericTypes represents types that support comparison operators
type NumericTypes interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// BasicTypes represents types that are basic types
type BasicTypes interface {
	~string | ~bool | NumericTypes
}

// Condition is a condition that can be used to validate a value.
type Condition[T any] interface {
	Validate(T) error
}

// ParameterProvider is a provider that can be used to get a parameter value.
type ParameterProvider[T any] interface {
	GetParamValue(context.Context, plow.Request) (T, plow.Response)
}

// OptionalParameterProvider extends ParameterProvider with Require flag for optional checking
type OptionalParameterProvider[T any] interface {
	ParameterProvider[T]
	Require() OptionalParameterProvider[T]
}

// ErrorResponse returns a response with a bad request status code and the error message.
func ErrorResponse(format string, p ...any) plow.Response {
	body := errorResponse{
		Error: fmt.Sprintf(format, p...),
	}
	return plow.JsonResponse(specs.StatusCodeBadRequest, body)
}

type errorResponse struct {
	Error string `json:"error"`
}

func bitSizeNum[T any](v T) int {
	return int(unsafe.Sizeof(v) * 8)
}
