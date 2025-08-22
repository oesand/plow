package mux

import (
	"errors"
	"fmt"
	"github.com/oesand/plow"
	"github.com/oesand/plow/internal/routing"
	"github.com/oesand/plow/specs"
	"iter"
	"slices"
	"strings"
)

func newRoute(method specs.HttpMethod, pattern string, handler plow.Handler, flags []any) (*route, error) {
	if pattern == "" {
		return nil, errors.New("plow: route pattern must have at least one character")
	}
	if pattern[0] != '/' {
		return nil, fmt.Errorf("plow: route pattern must starts with '/': %s", pattern)
	}

	if !method.IsValid() {
		return nil, fmt.Errorf("plow: invalid http method: %s", pattern)
	}
	if handler == nil {
		return nil, fmt.Errorf("plow: nil handler: %s", pattern)
	}

	if len(pattern) > 2 {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	routePattern, err := routing.ParseRoutePattern(pattern)
	if err != nil {
		return nil, err
	}

	return &route{
		RoutePattern: *routePattern,
		method:       method,
		handler:      handler,
		flags:        flags,
	}, nil
}

type route struct {
	routing.RoutePattern
	method  specs.HttpMethod
	handler plow.Handler
	flags   []any
}

func (rb *route) Method() specs.HttpMethod {
	return rb.method
}

func (rb *route) Pattern() string {
	return rb.RoutePattern.Original
}

func (rb *route) Handler() plow.Handler {
	return rb.handler
}

func (rb *route) Flags() iter.Seq[any] {
	return slices.Values(rb.flags)
}
