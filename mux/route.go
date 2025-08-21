package mux

import (
	"errors"
	"github.com/oesand/plow"
	"github.com/oesand/plow/internal/routing"
	"github.com/oesand/plow/specs"
	"iter"
	"slices"
)

func newRoute(method specs.HttpMethod, path string, handler plow.Handler, flags []any) (*route, error) {
	if !method.IsValid() {
		return nil, errors.New("plow: invalid http method")
	}
	if handler == nil {
		return nil, errors.New("plow: nil handler")
	}
	pattern, err := routing.parseRouteTemplate(path)
	if err != nil {
		return nil, err
	}
	return &route{
		routePattern: *pattern,
		method:       method,
		handler:      handler,
		flags:        flags,
	}, nil
}

type route struct {
	routing.routePattern
	method  specs.HttpMethod
	handler plow.Handler
	flags   []any
}

func (rb *route) Method() specs.HttpMethod {
	return rb.method
}

func (rb *route) Path() string {
	return rb.routePattern.Template
}

func (rb *route) Handler() plow.Handler {
	return rb.handler
}

func (rb *route) Flags() iter.Seq[any] {
	return slices.Values(rb.flags)
}
