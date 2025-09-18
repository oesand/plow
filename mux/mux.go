package mux

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"sort"
	"sync"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// New creates a new Mux instance with optional initial routers.
// If routers are provided, they will be included in the new mux.
// This is the primary constructor for creating new multiplexer instances.
func New(routers ...RouterBuilder) Mux {
	mx := &mux{}
	for _, router := range routers {
		mx.Include(router)
	}
	return mx
}

type mux struct {
	routes          map[specs.HttpMethod][]*route
	middlewares     []Middleware
	notFoundHandler plow.Handler

	mu sync.RWMutex
}

func (mx *mux) Use(md Middleware) Mux {
	if md == nil {
		panic("plow: nil Middleware")
	}
	mx.mu.Lock()
	defer mx.mu.Unlock()

	mx.middlewares = append(mx.middlewares, md)
	return mx
}

func (mx *mux) Route(method specs.HttpMethod, path string, handler plow.Handler, flags ...any) Mux {
	rt, err := newRoute(method, path, handler, flags)
	if err != nil {
		panic(err)
	}
	mx.mu.Lock()
	defer mx.mu.Unlock()

	if mx.routes == nil {
		mx.routes = make(map[specs.HttpMethod][]*route)
	}

	routes := mx.routes[method]
	routes = append(routes, rt)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Depth == routes[j].Depth {
			return len(routes[i].ParamNames) < len(routes[j].ParamNames)
		}
		return routes[i].Depth > routes[j].Depth
	})
	mx.routes[method] = routes
	return mx
}

func (mx *mux) Include(rb RouterBuilder) Mux {
	if rb == nil {
		panic("plow: nil RouterBuilder")
	}
	for rt := range rb.Routes() {
		mx.Route(rt.Method(), rt.Pattern(), rt.Handler(), slices.Collect(rt.Flags())...)
	}
	return mx
}

func (mx *mux) NotFoundHandler(handler plow.Handler) Mux {
	mx.mu.Lock()
	defer mx.mu.Unlock()

	mx.notFoundHandler = handler
	return mx
}

func (mx *mux) Routes() iter.Seq[MuxRoute] {
	return func(yield func(MuxRoute) bool) {
		mx.mu.RLock()
		defer mx.mu.RUnlock()

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
	mx.mu.RLock()
	defer mx.mu.RUnlock()

	if len(mx.middlewares) > 0 {
		nextMd, stop := iter.Pull(slices.Values(mx.middlewares))
		defer stop()

		var nextFunc NextFunc
		nextFunc = func(ctx context.Context) plow.Response {
			if md, ok := nextMd(); ok {
				return md.Intercept(ctx, request, nextFunc)
			}
			return mx.handle(ctx, request)
		}

		return nextFunc(ctx)
	}

	return mx.handle(ctx, request)
}

func (mx *mux) handle(ctx context.Context, request plow.Request) plow.Response {
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
