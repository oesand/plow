package mux

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
)

type RouterBuilder interface {
	AddRoute(method specs.HttpMethod, path string, handler plow.Handler) RouteBuilder
	Routes() iter.Seq[Route]
}

type RouteBuilder interface {
	AddFlag(flags ...any) RouteBuilder
}

type Route interface {
	Method() specs.HttpMethod
	Path() string
	Handler() plow.Handler
	Flags() iter.Seq[any]
}

type Mux interface {
	plow.Handler
	NotFoundHandler(plow.Handler) Mux
	Include(RouterBuilder) Mux
	Add(method specs.HttpMethod, path string, handler plow.Handler, flags ...any) Mux

	Routes() iter.Seq[MuxRoute]
}

type MuxRoute interface {
	Route
	Match(path string) (bool, iter.Seq2[string, string])
}
