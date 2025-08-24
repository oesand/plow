package mux

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// Router creates a new RouterBuilder with optional configuration functions.
// The configuration functions are applied to the router during creation,
// allowing for declarative router setup.
func Router(configure ...func(router RouterBuilder)) RouterBuilder {
	rt := &routerBuilder{}

	for _, conf := range configure {
		conf(rt)
	}

	return rt
}

// PrefixRouter creates a new RouterBuilder with a URL prefix that will be prepended
// to all routes added to this router. This is useful for grouping related routes
// under a common path prefix.
//
// Panics if the prefix is less than 2 characters or doesn't start with '/'.
func PrefixRouter(prefix string, configure ...func(router RouterBuilder)) RouterBuilder {
	if len(prefix) < 2 {
		panic("plow: router prefix must have at least two characters")
	}
	if prefix[0] != '/' {
		panic(fmt.Sprintf("plow: router prefix must starts with '/': %s", prefix))
	}

	rt := &routerBuilder{
		prefix: strings.TrimSuffix(prefix, "/"),
	}

	for _, conf := range configure {
		conf(rt)
	}

	return rt
}

type routerBuilder struct {
	prefix string
	routes []*routeBuilder
}

func (rb *routerBuilder) Route(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) RouteBuilder {
	if pattern == "" {
		panic("plow: route pattern must have at least one character")
	}
	if pattern[0] != '/' {
		panic(fmt.Sprintf("plow: route pattern must starts with '/': %s", pattern))
	}

	if !method.IsValid() {
		panic(fmt.Sprintf("plow: invalid http method: %s", pattern))
	}
	if handler == nil {
		panic(fmt.Sprintf("plow: nil handler: %s", pattern))
	}

	if len(pattern) > 2 {
		pattern = strings.TrimSuffix(pattern, "/")
	}

	if rb.prefix != "" {
		if pattern == "/" {
			pattern = rb.prefix
		} else {
			pattern = rb.prefix + pattern
		}
	}

	builder := &routeBuilder{
		method:  method,
		pattern: pattern,
		handler: handler,
		flags:   flags,
	}
	rb.routes = append(rb.routes, builder)
	return builder
}

func (rb *routerBuilder) Include(other RouterBuilder) RouterBuilder {
	if other == rb {
		panic("plow: router builder cannot include self")
	}
	for rt := range other.Routes() {
		rb.Route(rt.Method(), rt.Pattern(), rt.Handler(), slices.Collect(rt.Flags())...)
	}
	return rb
}

func (rb *routerBuilder) Routes() iter.Seq[Route] {
	return func(yield func(Route) bool) {
		for _, rt := range rb.routes {
			if !yield(rt) {
				break
			}
		}
	}
}

type routeBuilder struct {
	method  specs.HttpMethod
	pattern string
	handler plow.Handler
	flags   []any
}

func (rb *routeBuilder) Method() specs.HttpMethod {
	return rb.method
}

func (rb *routeBuilder) Pattern() string {
	return rb.pattern
}

func (rb *routeBuilder) Handler() plow.Handler {
	return rb.handler
}

func (rb *routeBuilder) Flags() iter.Seq[any] {
	return slices.Values(rb.flags)
}

func (rb *routeBuilder) AddFlag(flags ...any) RouteBuilder {
	rb.flags = append(rb.flags, flags...)
	return rb
}
