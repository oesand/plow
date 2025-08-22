package mux

import (
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"reflect"
	"slices"
	"testing"
)

func TestRouter(t *testing.T) {
	t.Run("Route", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			method  specs.HttpMethod
			pattern string
			flags   []any
		}

		router := Router(func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/", handler)
			router.Route(specs.HttpMethodPost, "/post", handler)
			router.Route(specs.HttpMethodPut, "/put/", handler)
			router.Route(specs.HttpMethodDelete, "/delete/123/", handler)

			router.Route(specs.HttpMethodGet, "/one", handler, "hello", "world")
			router.Route(specs.HttpMethodPost, "/two/", handler).AddFlag("flag")
			router.Route(specs.HttpMethodGet, "/third/", handler, "hello").AddFlag("flag", "two")
		})

		expectedRoutes := []routeItem{
			{method: specs.HttpMethodGet, pattern: "/", flags: []any{}},
			{method: specs.HttpMethodPost, pattern: "/post", flags: []any{}},
			{method: specs.HttpMethodPut, pattern: "/put", flags: []any{}},
			{method: specs.HttpMethodDelete, pattern: "/delete/123", flags: []any{}},

			{method: specs.HttpMethodGet, pattern: "/one", flags: []any{"hello", "world"}},
			{method: specs.HttpMethodPost, pattern: "/two", flags: []any{"flag"}},
			{method: specs.HttpMethodGet, pattern: "/third", flags: []any{"hello", "flag", "two"}},
		}

		var i int
		for rt := range router.Routes() {
			want := expectedRoutes[i]
			if !reflect.DeepEqual(rt.Method(), want.method) {
				t.Errorf("Router().Method() = %v, want %v", rt.Method(), want.method)
			}
			if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
				t.Errorf("Router().Pattern() = %v, want %v", rt.Pattern(), want.pattern)
			}
			if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
				t.Errorf("Router().Flags() = %v, want %v", flags, want.flags)
			}
			i++
		}

		if i != len(expectedRoutes) {
			t.Errorf("Router().Routes().Len = %v, want %v", i, len(expectedRoutes))
		}
	})

	t.Run("Include", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			method  specs.HttpMethod
			pattern string
			flags   []any
		}

		otherRoute := Router(func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/", handler)
			router.Route(specs.HttpMethodPost, "/post", handler)
			router.Route(specs.HttpMethodPut, "/put/", handler)
			router.Route(specs.HttpMethodDelete, "/delete/123/", handler)

			router.Route(specs.HttpMethodGet, "/one", handler, "hello", "world")
			router.Route(specs.HttpMethodPost, "/two/", handler).AddFlag("flag")
			router.Route(specs.HttpMethodGet, "/third/", handler, "hello").AddFlag("flag", "two")
		})

		router := Router().Include(otherRoute)

		expectedRoutes := []routeItem{
			{method: specs.HttpMethodGet, pattern: "/", flags: []any{}},
			{method: specs.HttpMethodPost, pattern: "/post", flags: []any{}},
			{method: specs.HttpMethodPut, pattern: "/put", flags: []any{}},
			{method: specs.HttpMethodDelete, pattern: "/delete/123", flags: []any{}},

			{method: specs.HttpMethodGet, pattern: "/one", flags: []any{"hello", "world"}},
			{method: specs.HttpMethodPost, pattern: "/two", flags: []any{"flag"}},
			{method: specs.HttpMethodGet, pattern: "/third", flags: []any{"hello", "flag", "two"}},
		}

		var i int
		for rt := range router.Routes() {
			want := expectedRoutes[i]
			if !reflect.DeepEqual(rt.Method(), want.method) {
				t.Errorf("Router().Method() = %v, want %v", rt.Method(), want.method)
			}
			if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
				t.Errorf("Router().Pattern() = %v, want %v", rt.Pattern(), want.pattern)
			}
			if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
				t.Errorf("Router().Flags() = %v, want %v", flags, want.flags)
			}
			i++
		}

		if i != len(expectedRoutes) {
			t.Errorf("Router().Routes().Len = %v, want %v", i, len(expectedRoutes))
		}
	})
}

func TestPrefixRouter(t *testing.T) {
	t.Run("Route", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			method  specs.HttpMethod
			pattern string
			flags   []any
		}

		router := PrefixRouter("/prefix/", func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/", handler)
			router.Route(specs.HttpMethodPost, "/post", handler)
			router.Route(specs.HttpMethodPut, "/put/", handler)
			router.Route(specs.HttpMethodDelete, "/delete/123/", handler)

			router.Route(specs.HttpMethodGet, "/one", handler, "hello", "world")
			router.Route(specs.HttpMethodPost, "/two/", handler).AddFlag("flag")
			router.Route(specs.HttpMethodGet, "/third/", handler, "hello").AddFlag("flag", "two")
		})

		expectedRoutes := []routeItem{
			{method: specs.HttpMethodGet, pattern: "/prefix", flags: []any{}},
			{method: specs.HttpMethodPost, pattern: "/prefix/post", flags: []any{}},
			{method: specs.HttpMethodPut, pattern: "/prefix/put", flags: []any{}},
			{method: specs.HttpMethodDelete, pattern: "/prefix/delete/123", flags: []any{}},

			{method: specs.HttpMethodGet, pattern: "/prefix/one", flags: []any{"hello", "world"}},
			{method: specs.HttpMethodPost, pattern: "/prefix/two", flags: []any{"flag"}},
			{method: specs.HttpMethodGet, pattern: "/prefix/third", flags: []any{"hello", "flag", "two"}},
		}

		var i int
		for rt := range router.Routes() {
			want := expectedRoutes[i]
			if !reflect.DeepEqual(rt.Method(), want.method) {
				t.Errorf("Router().Method() = %v, want %v", rt.Method(), want.method)
			}
			if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
				t.Errorf("Router().Pattern() = %v, want %v", rt.Pattern(), want.pattern)
			}
			if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
				t.Errorf("Router().Flags() = %v, want %v", flags, want.flags)
			}
			i++
		}

		if i != len(expectedRoutes) {
			t.Errorf("Router().Routes().Len = %v, want %v", i, len(expectedRoutes))
		}
	})

	t.Run("Include", func(t *testing.T) {
		handler := plow.HandlerFunc(nil)
		type routeItem struct {
			method  specs.HttpMethod
			pattern string
			flags   []any
		}

		otherRouter := PrefixRouter("/router/v1/", func(router RouterBuilder) {
			router.Route(specs.HttpMethodGet, "/", handler)
			router.Route(specs.HttpMethodPost, "/post", handler)
			router.Route(specs.HttpMethodPut, "/put/", handler)
			router.Route(specs.HttpMethodDelete, "/delete/123/", handler)

			router.Route(specs.HttpMethodGet, "/one", handler, "hello", "world")
			router.Route(specs.HttpMethodPost, "/two/", handler).AddFlag("flag")
			router.Route(specs.HttpMethodGet, "/third/", handler, "hello").AddFlag("flag", "two")
		})

		router := PrefixRouter("/api").Include(otherRouter)

		expectedRoutes := []routeItem{
			{method: specs.HttpMethodGet, pattern: "/api/router/v1", flags: []any{}},
			{method: specs.HttpMethodPost, pattern: "/api/router/v1/post", flags: []any{}},
			{method: specs.HttpMethodPut, pattern: "/api/router/v1/put", flags: []any{}},
			{method: specs.HttpMethodDelete, pattern: "/api/router/v1/delete/123", flags: []any{}},

			{method: specs.HttpMethodGet, pattern: "/api/router/v1/one", flags: []any{"hello", "world"}},
			{method: specs.HttpMethodPost, pattern: "/api/router/v1/two", flags: []any{"flag"}},
			{method: specs.HttpMethodGet, pattern: "/api/router/v1/third", flags: []any{"hello", "flag", "two"}},
		}

		var i int
		for rt := range router.Routes() {
			want := expectedRoutes[i]
			if !reflect.DeepEqual(rt.Method(), want.method) {
				t.Errorf("Router().Method() = %v, want %v", rt.Method(), want.method)
			}
			if !reflect.DeepEqual(rt.Pattern(), want.pattern) {
				t.Errorf("Router().Pattern() = %v, want %v", rt.Pattern(), want.pattern)
			}
			if flags := slices.Collect(rt.Flags()); !slices.Equal(flags, want.flags) {
				t.Errorf("Router().Flags() = %v, want %v", flags, want.flags)
			}
			i++
		}

		if i != len(expectedRoutes) {
			t.Errorf("Router().Routes().Len = %v, want %v", i, len(expectedRoutes))
		}
	})
}
