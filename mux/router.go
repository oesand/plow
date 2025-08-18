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
		panic("router prefix must have at least two characters")
	}
	if prefix[0] != '/' {
		panic(fmt.Sprintf("router prefix must starts with '/': %s", prefix))
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

func (rb *routerBuilder) AddRoute(method specs.HttpMethod, path string, handler plow.Handler) RouteBuilder {
	if !method.IsValid() {
		panic("invalid http method")
	}
	if handler == nil {
		panic("nil handler")
	}
	if path == "" {
		panic("router prefix must have at least one character")
	}
	if path[0] != '/' {
		panic(fmt.Sprintf("path must starts with '/': %s", path))
	}
	if len(path) > 2 {
		path = strings.TrimSuffix(path, "/")
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

func (rb *routeBuilder) Flags() iter.Seq[any] {
	return slices.Values(rb.flags)
}

func (rb *routeBuilder) AddFlag(flags ...any) RouteBuilder {
	copy(rb.flags, flags)
	return rb
}
