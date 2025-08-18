package mux

import (
	"context"
	"fmt"
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"iter"
	"sort"
)

func Make(routers ...RouterBuilder) Mux {
	mx := &mux{}
	for _, router := range routers {
		mx.Include(router)
	}
	return mx
}

type mux struct {
	routes          map[specs.HttpMethod][]*route
	notFoundHandler plow.Handler
}

func (mx *mux) Add(method specs.HttpMethod, path string, handler plow.Handler, flags ...any) Mux {
	rt, err := newRoute(method, path, handler, flags)
	if err != nil {
		panic(fmt.Errorf("cannot add route <%s>'%s': %s", method, path, err))
	}

	if mx.routes == nil {
		mx.routes = make(map[specs.HttpMethod][]*route)
	}

	mx.routes[method] = append(mx.routes[method], rt)
	routes := mx.routes[method]
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path() < routes[j].Path()
	})
	return mx
}

func (mx *mux) Include(rb RouterBuilder) Mux {
	for rt := range rb.Routes() {
		mx.Add(rt.Method(), rt.Path(), rt.Handler(), rt.Flags())
	}
	return mx
}

func (mx *mux) NotFoundHandler(handler plow.Handler) Mux {
	mx.notFoundHandler = handler
	return mx
}

func (mx *mux) Routes() iter.Seq[MuxRoute] {
	return func(yield func(MuxRoute) bool) {
		for _, routes := range mx.routes {
			for _, rt := range routes {
				if !yield(rt) {
					break
				}
			}
		}
	}
}

func (mx *mux) Handle(ctx context.Context, request plow.Request) plow.Response {
	if mx.routes != nil {
		url := request.Url()
		routes := mx.routes[request.Method()]
		for _, rt := range routes {
			if ok, params := rt.Match(url.Path); ok {
				for key, value := range params {
					if url.Query == nil {
						url.Query = make(specs.Query)
					}
					url.Query[key] = value
				}
				return rt.Handler().Handle(ctx, request)
			}
		}
	}

	if handler := mx.notFoundHandler; handler != nil {
		return handler.Handle(ctx, request)
	}

	return plow.TextResponse(specs.StatusCodeNotFound, specs.ContentTypePlain,
		fmt.Sprintf("Not Found %s", request.Url().Path))
}
