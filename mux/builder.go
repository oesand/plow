package mux

import (
	"fmt"
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
	"slices"
	"strings"
)

func Router(configure ...func(router RouterBuilder)) RouterBuilder {
	return newRouter(configure)
}

func PrefixRouter(prefix string, configure ...func(router RouterBuilder)) RouterBuilder {
	if len(prefix) < 2 {
		panic("plow: router prefix must have at least two characters")
	}
	if prefix[0] != '/' {
		panic(fmt.Sprintf("plow: router prefix must starts with '/': %s", prefix))
	}

	rt := newRouter(configure)
	rt.prefix = strings.TrimSuffix(prefix, "/")
	return rt
}

func newRouter(configure []func(router RouterBuilder)) *routerBuilder {
	rt := &routerBuilder{}

	for _, conf := range configure {
		conf(rt)
	}

	return rt
}

type routerBuilder struct {
	prefix string
	routes []*routeBuilder
}

func (rb *routerBuilder) AddRoute(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) RouteBuilder {
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

	builder := &routeBuilder{
		method:  method,
		path:    rb.prefix + pattern,
		handler: handler,
		flags:   flags,
	}
	rb.routes = append(rb.routes, builder)
	return builder
}

func (rb *routerBuilder) Include(other RouterBuilder) RouterBuilder {
	for rt := range other.Routes() {
		rb.AddRoute(rt.Method(), rt.Path(), rt.Handler()).AddFlag(rt.Flags())
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
	path    string
	handler plow.Handler
	flags   []any
}

func (rb *routeBuilder) Method() specs.HttpMethod {
	return rb.method
}

func (rb *routeBuilder) Path() string {
	return rb.path
}

func (rb *routeBuilder) Handler() plow.Handler {
	return rb.handler
}

func (rb *routeBuilder) Flags() iter.Seq[any] {
	return slices.Values(rb.flags)
}

func (rb *routeBuilder) AddFlag(flags ...any) RouteBuilder {
	copy(rb.flags, flags)
	return rb
}
