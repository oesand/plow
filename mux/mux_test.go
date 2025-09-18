package mux

import (
	"context"
	"github.com/oesand/plow"
	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
	"reflect"
	"slices"
	"sync/atomic"
	"testing"
)

func TestMux(t *testing.T) {
	t.Run("Route", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			method  specs.HttpMethod
			pattern string
			flags   []any
		}

		mx := New().(*mux)
		mx.Route(specs.HttpMethodGet, "/", handler)
		mx.Route(specs.HttpMethodPost, "/post", handler)
		mx.Route(specs.HttpMethodPut, "/put/", handler)
		mx.Route(specs.HttpMethodDelete, "/delete/123/", handler)

		mx.Route(specs.HttpMethodGet, "/api/{prm}/", handler)
		mx.Route(specs.HttpMethodGet, "/api/flow/", handler)
		mx.Route(specs.HttpMethodGet, "/api/inner/{prm}/", handler)
		mx.Route(specs.HttpMethodGet, "/api/flagged", handler, "hello", "world")
		mx.Route(specs.HttpMethodGet, "/api/inner/flow/", handler)

		expectedRoutes := map[specs.HttpMethod][]routeItem{
			specs.HttpMethodGet: {
				{pattern: "/api/inner/flow", flags: []any{}},
				{pattern: "/api/inner/{prm}", flags: []any{}},
				{pattern: "/api/flow", flags: []any{}},
				{pattern: "/api/flagged", flags: []any{"hello", "world"}},
				{pattern: "/api/{prm}", flags: []any{}},
				{pattern: "/", flags: []any{}},
			},
			specs.HttpMethodPost: {
				{pattern: "/post", flags: []any{}},
			},
			specs.HttpMethodPut: {
				{pattern: "/put", flags: []any{}},
			},
			specs.HttpMethodDelete: {
				{pattern: "/delete/123", flags: []any{}},
			},
		}

		for method, exroutes := range expectedRoutes {
			var i int
			for _, rt := range mx.routes[method] {
				want := exroutes[i]
				if !reflect.DeepEqual(rt.Method(), method) {
					t.Errorf("Mux.Method() = %v, want %v", rt.Method(), method)
				}
				if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
					t.Errorf("Mux.Pattern() = %v, want %v", rt.Pattern(), want.pattern)
				}
				if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
					t.Errorf("Mux.Flags() = %v, want %v", flags, want.flags)
				}
				i++
			}
			if i != len(exroutes) {
				t.Errorf("Mux.Routes(%s).Len = %v, want %v", method, i, len(exroutes))
			}
		}
	})

	t.Run("Include", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			pattern string
			flags   []any
		}

		routerOne := Router(func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/", handler)
			router.Route(specs.HttpMethodPost, "/post", handler)
		})

		routerTwo := PrefixRouter("/api/route/v1/", func(router RouterBuilder) {
			router.Route(specs.HttpMethodPut, "/put/", handler)
			router.Route(specs.HttpMethodDelete, "/delete/123/", handler)
		})

		routerThree := PrefixRouter("/api/v2/", func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/{prm}/", handler)
			router.Route(specs.HttpMethodGet, "/flow/", handler)
			router.Route(specs.HttpMethodGet, "/inner/{prm}/", handler)
			router.Route(specs.HttpMethodGet, "/flagged", handler, "hello", "world")
			router.Route(specs.HttpMethodGet, "/inner/flow/", handler)
		})

		mx := New(routerOne, routerTwo).(*mux)
		mx.Include(routerThree)

		expectedRoutes := map[specs.HttpMethod][]routeItem{
			specs.HttpMethodGet: {
				{pattern: "/api/v2/inner/flow", flags: []any{}},
				{pattern: "/api/v2/inner/{prm}", flags: []any{}},
				{pattern: "/api/v2/flow", flags: []any{}},
				{pattern: "/api/v2/flagged", flags: []any{"hello", "world"}},
				{pattern: "/api/v2/{prm}", flags: []any{}},
				{pattern: "/", flags: []any{}},
			},
			specs.HttpMethodPost: {
				{pattern: "/post", flags: []any{}},
			},
			specs.HttpMethodPut: {
				{pattern: "/api/route/v1/put", flags: []any{}},
			},
			specs.HttpMethodDelete: {
				{pattern: "/api/route/v1/delete/123", flags: []any{}},
			},
		}

		for method, exroutes := range expectedRoutes {
			var i int
			for _, rt := range mx.routes[method] {
				want := exroutes[i]
				if !reflect.DeepEqual(rt.Method(), method) {
					t.Errorf("Mux.Method() = %v, want %v", rt.Method(), method)
				}
				if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
					t.Errorf("Mux.Pattern() = %v, want %v", rt.Pattern(), want.pattern)
				}
				if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
					t.Errorf("Mux.Flags() = %v, want %v", flags, want.flags)
				}
				i++
			}
			if i != len(exroutes) {
				t.Errorf("Mux.Routes(%s).Len = %v, want %v", method, i, len(exroutes))
			}
		}
	})

	t.Run("Handle", func(t *testing.T) {
		var firstMiddleware atomic.Int32
		var secondMiddleware atomic.Int32
		var thirdMiddleware atomic.Int32
		var visitGet atomic.Bool
		var visitPost atomic.Bool
		var visitPattern atomic.Bool
		var visitNotFound atomic.Bool

		mx := New()

		mx.Use(MiddlewareFunc(func(ctx context.Context, request plow.Request, next NextFunc) plow.Response {
			firstMiddleware.Add(1)
			return next(ctx)
		}))

		mx.Use(MiddlewareFunc(func(ctx context.Context, request plow.Request, next NextFunc) plow.Response {
			secondMiddleware.Add(1)
			return next(ctx)
		}))

		mx.Use(MiddlewareFunc(func(ctx context.Context, request plow.Request, next NextFunc) plow.Response {
			thirdMiddleware.Add(1)
			return next(ctx)
		}))

		mx.NotFoundHandler(plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
			if visitNotFound.Load() {
				t.Errorf("twice visit not found")
			}
			visitNotFound.Store(true)
			return nil
		}))

		mx.Route(specs.HttpMethodGet, "/", plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
			if visitGet.Load() {
				t.Errorf("twice visit handle")
			}
			visitGet.Store(true)
			return nil
		}))

		mx.Route(specs.HttpMethodPost, "/post", plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
			if visitPost.Load() {
				t.Errorf("twice visit handle")
			}
			visitPost.Store(true)
			return nil
		}))

		mx.Route(specs.HttpMethodPut, "/put/{id}", plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
			if visitPattern.Load() {
				t.Errorf("twice visit handle")
			}
			visitPattern.Store(true)

			if request.Url().Query["id"] != "expected790" {
				t.Errorf("wrong path parameter: %v", request.Url().Query)
			}
			return nil
		}))

		ctx := context.Background()

		// Check GET "/"
		t.Run("Check GET", func(t *testing.T) {
			mx.Handle(ctx, mock.DefaultRequest().Url(specs.MustParseUrl("")).Request())
			if !visitGet.Load() {
				t.Errorf("not visited")
			}
			visitGet.Store(false)

			mx.Handle(ctx, mock.DefaultRequest().Url(specs.MustParseUrl("/")).Request())
			if !visitGet.Load() {
				t.Errorf("not visited")
			}
			visitGet.Store(false)
		})

		// Check POST "/post"
		t.Run("Check POST", func(t *testing.T) {
			mx.Handle(ctx, mock.DefaultRequest().Method(specs.HttpMethodPost).Url(specs.MustParseUrl("/post")).Request())
			if !visitPost.Load() {
				t.Errorf("not visited")
			}
			visitPost.Store(false)
		})

		// Check PUT "/put/expected790"
		t.Run("Check PUT", func(t *testing.T) {
			mx.Handle(ctx, mock.DefaultRequest().Method(specs.HttpMethodPut).Url(specs.MustParseUrl("/put/expected790")).Request())
			if !visitPattern.Load() {
				t.Errorf("not visited")
			}
			visitPattern.Store(false)

			mx.Handle(ctx, mock.DefaultRequest().Method(specs.HttpMethodPut).Url(specs.MustParseUrl("/put/expected790/")).Request())
			if !visitPattern.Load() {
				t.Errorf("not visited")
			}
			visitPattern.Store(false)
		})

		// Check NotFound
		t.Run("Check NotFound", func(t *testing.T) {
			mx.Handle(ctx, mock.DefaultRequest().Method(specs.HttpMethodGet).Url(specs.MustParseUrl("/nf")).Request())
			if !visitNotFound.Load() {
				t.Errorf("not visited")
			}
			visitNotFound.Store(false)

			mx.Handle(ctx, mock.DefaultRequest().Method(specs.HttpMethodDelete).Url(specs.MustParseUrl("/")).Request())
			if !visitNotFound.Load() {
				t.Errorf("not visited")
			}
			visitNotFound.Store(false)
		})

		// Check Middleware
		t.Run("Check Middleware", func(t *testing.T) {
			firstMiddleware.Store(0)
			secondMiddleware.Store(0)
			thirdMiddleware.Store(0)

			mx.Handle(ctx, mock.DefaultRequest().Url(specs.MustParseUrl("")).Request())
			if !visitGet.Load() {
				t.Errorf("not visited")
			}
			visitGet.Store(false)

			mx.Handle(ctx, mock.DefaultRequest().Url(specs.MustParseUrl("/")).Request())
			if !visitGet.Load() {
				t.Errorf("not visited")
			}
			visitGet.Store(false)

			if firstMiddleware.Load() != 2 ||
				firstMiddleware.Load() != secondMiddleware.Load() ||
				secondMiddleware.Load() != thirdMiddleware.Load() {

				t.Errorf("Invalid middleware visits: %d, %d, %d", firstMiddleware.Load(), secondMiddleware.Load(), thirdMiddleware.Load())
			}
		})
	})
}
