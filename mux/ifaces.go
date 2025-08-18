package mux

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
)

type PathKind int

const (
	PathRaw PathKind = iota
	PathRegex
)

type RouterBuilder interface {
	AddRoute(method specs.HttpMethod, path string, handler plow.Handler) RouteBuilder
	Routes() iter.Seq[Route]
}

type Route interface {
	Method() specs.HttpMethod
	Path() string
	Handler() plow.Handler
	Flags() []any
}

type RouteBuilder interface {
	AddFlag(flags ...any) RouteBuilder
}

type Mux interface {
	plow.Handler
	Routes() iter.Seq[Route]
}
