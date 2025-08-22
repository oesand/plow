package prm

import (
	"context"

	"github.com/oesand/plow"
)

// RequestParam creates a new ParameterProvider for extracting the raw request object.
// This is useful when you need access to the complete request object in your handler.
func RequestParam() ParameterProvider[plow.Request] {
	return &requestParameter{}
}

type requestParameter struct{}

func (rp *requestParameter) GetParamValue(_ context.Context, req plow.Request) (plow.Request, plow.Response) {
	return req, nil
}
