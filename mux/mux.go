package mux

import (
	"context"
	"github.com/oesand/plow"
	"iter"
)

func Make(routers ...RouterBuilder) Mux {
	return &mux{}
}

type mux struct {
	routes []route
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

func (m *mux) Handle(ctx context.Context, request plow.Request) plow.Response {

	return nil
}
