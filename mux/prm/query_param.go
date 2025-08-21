package prm

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/oesand/plow"
)

// QueryParam creates a new query parameter.
func QueryParam[T BasicTypes](name string, conditions ...Condition[T]) OptionalParameterProvider[T] {
	return &queryParameter[T]{
		name:       name,
		conditions: conditions,
	}
}

type queryParameter[T BasicTypes] struct {
	name       string
	required   bool
	conditions []Condition[T]
}

func (qp *queryParameter[T]) Require() OptionalParameterProvider[T] {
	qp.required = true
	return qp
}

func (qp *queryParameter[T]) GetParamValue(_ context.Context, req plow.Request) (T, plow.Response) {
	var str string
	if req.Url().Query.Any() {
		str, _ = req.Url().Query[qp.name]
	}

	var val T
	var resp plow.Response
	if str == "" {
		if qp.required {
			resp = ErrorResponse("query parameter '%s' is required", qp.name)
		}
		return val, resp
	}

	switch any(val).(type) {
	case string:
		val = any(str).(T)
	case bool:
		bv, err := strconv.ParseBool(str)
		if err != nil {
			resp = ErrorResponse("query parameter '%s' must be bool", qp.name)
			break
		}
		val = any(bv).(T)
	case uint, uint8, uint16, uint32, uint64:
		bitSize := bitSizeNum(val)
		uiv, err := strconv.ParseUint(str, 10, bitSize)
		if err != nil {
			resp = ErrorResponse("query parameter '%s' must be integer", qp.name)
			break
		}
		val = any(uiv).(T)
	case int, int8, int16, int32, int64:
		bitSize := bitSizeNum(val)
		iv, err := strconv.ParseInt(str, 10, bitSize)
		if err != nil {
			resp = ErrorResponse("query parameter '%s' must be integer", qp.name)
			break
		}
		val = any(iv).(T)
	case float32, float64:
		bitSize := bitSizeNum(val)
		iv, err := strconv.ParseFloat(str, bitSize)
		if err != nil {
			resp = ErrorResponse("query parameter '%s' must be float", qp.name)
			break
		}
		val = any(iv).(T)
	default:
		panic(fmt.Sprintf("plow: unknown type: %s", reflect.TypeFor[T]().String()))
	}

	for _, condition := range qp.conditions {
		if err := condition.Validate(val); err != nil {
			resp = ErrorResponse("query parameter '%s' is invalid: %s", qp.name, err)
			break
		}
	}
	return val, resp
}
