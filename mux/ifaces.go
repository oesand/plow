package mux

import (
	"context"
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
)

type NextFunc func(ctx context.Context) plow.Response
type Middleware func(ctx context.Context, request plow.Request, next NextFunc) plow.Response

type RouterBuilder interface {
	AddRoute(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) RouteBuilder
	Include(RouterBuilder) RouterBuilder
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
	Use(middleware Middleware) Mux

	Add(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) Mux
	Include(router RouterBuilder) Mux
	NotFoundHandler(handler plow.Handler) Mux

	Routes() iter.Seq[MuxRoute]
}

type MuxRoute interface {
	Route
	Match(path string) (bool, iter.Seq2[string, string])
}
