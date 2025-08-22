package mux

import (
	"context"
	"iter"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// NextFunc represents a function that continues the middleware chain execution.
// It takes a context and returns the next response in the chain.
type NextFunc func(ctx context.Context) plow.Response

// Middleware represents a function that can intercept and process HTTP requests
// before they reach the final handler. It can modify the request, execute
// additional logic, or short-circuit the request processing.
type Middleware func(ctx context.Context, request plow.Request, next NextFunc) plow.Response

// RouterBuilder provides an interface for building and configuring route collections.
// It allows adding routes, including other routers, and retrieving all configured routes.
type RouterBuilder interface {
	// Route adds a new route to the router with the specified HTTP method,
	// URL pattern, handler function, and optional flags.
	// Returns a RouteBuilder for further configuration of the route.
	//
	// Does not compile the pattern only stores it before placing it in the [Mux].
	// For more information about the available formats in Mux.Route
	Route(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) RouteBuilder

	// Include incorporates all routes from another RouterBuilder into this one.
	// This allows for modular router composition and reuse.
	Include(router RouterBuilder) RouterBuilder

	// Routes returns an iterator over all routes configured in this router.
	// The routes are returned as a sequence that can be iterated over.
	Routes() iter.Seq[Route]
}

// RouteBuilder provides an interface for configuring individual routes after they are created.
// It allows adding flags and other configuration options to a route.
type RouteBuilder interface {
	// AddFlag adds one or more flags to the route for additional configuration.
	// Flags can be used to modify route behavior or provide metadata.
	AddFlag(flags ...any) RouteBuilder
}

// Route represents a configured HTTP route with its method, path pattern, handler, and flags.
// It provides access to all the essential information about a route.
type Route interface {
	// Method returns the HTTP method (GET, POST, PUT, etc.) that this route handles.
	Method() specs.HttpMethod

	// Pattern returns the URL path pattern that this route matches against.
	Pattern() string

	// Handler returns the function that will process requests matching this route.
	Handler() plow.Handler

	// Flags returns an iterator over all flags associated with this route.
	// Flags can provide additional configuration or metadata for the route.
	Flags() iter.Seq[any]
}

// Mux is the main multiplexer interface that combines routing, middleware support,
// and request handling capabilities. It implements the plow.Handler interface
// and provides methods for building complex routing configurations.
type Mux interface {
	plow.Handler

	// Use adds a middleware function to the mux's middleware chain.
	// Middleware functions are executed in the order they are added.
	Use(middleware Middleware) Mux

	// Route creates a new route with the specified HTTP method, URL pattern,
	// handler function, and optional flags. Returns the mux for method chaining.
	//
	// Supported formats:
	//   - /users/{id} - string parameter
	//   - /files/{name}/raw/{id} - many parameters
	//   - /posts/{id:<regex pattern>} - regex parameter
	//   - /static/{*:.*} - wildcard parameter for "at the end" (also support regex)
	//
	// Everything outside {…} is safely regex-escaped
	// Trailing slash is ignored at compile-time; both /path and /path/ are accepted at match-time
	// Wildcard parameters (*) can match any characters including slashes
	Route(method specs.HttpMethod, pattern string, handler plow.Handler, flags ...any) Mux

	// Include incorporates all routes from a RouterBuilder into this mux.
	// This allows for modular router composition and reuse.
	Include(router RouterBuilder) Mux

	// NotFoundHandler sets a custom handler for requests that don't match any routes.
	// If not set, a default 404 response is returned.
	NotFoundHandler(handler plow.Handler) Mux

	// Routes returns an iterator over all routes configured in this mux.
	// The routes include both the route information and matching capabilities.
	Routes() iter.Seq[MuxRoute]
}

// MuxRoute extends the basic Route interface with path matching capabilities.
// It allows checking if a given path matches the route pattern and extracting
// path parameters.
type MuxRoute interface {
	Route

	// Match checks if the given path matches this route's pattern.
	// Returns true if there's a match, along with any extracted path parameters.
	// The parameters are returned as key-value pairs in a sequence.
	Match(path string) (bool, iter.Seq2[string, string])
}
