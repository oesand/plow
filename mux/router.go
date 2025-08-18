package mux

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
)

func Router(configure ...func(router RouterBuilder)) RouterBuilder {
	return newRouter(configure)
}

func PrefixRouter(prefix string, configure ...func(router RouterBuilder)) RouterBuilder {
	rt := newRouter(configure)
	rt.prefix = prefix
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

func (rb *routerBuilder) AddRoute(method specs.HttpMethod, path string, handler plow.Handler) RouteBuilder {
	if !method.IsValid() {
		panic("invalid http method")
	}
	if handler == nil {
		panic("nil handler")
	}
	builder := &routeBuilder{
		method:  method,
		path:    rb.prefix + path,
		handler: handler,
	}
	rb.routes = append(rb.routes, builder)
	return builder
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

func (rb *routeBuilder) Flags() []any {
	return rb.flags
}

func (rb *routeBuilder) AddFlag(flags ...any) RouteBuilder {
	copy(rb.flags, flags)
	return rb
}
