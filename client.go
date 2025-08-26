package plow

import (
	"context"
	"fmt"
	"github.com/oesand/plow/internal/catch"
	"github.com/oesand/plow/specs"
	"sync"
)

// DefaultClient factory for creating [Client]
// with optimal parameters for perfomance and safety
//
// Each call creates a new instance of [Client]
func DefaultClient() *Client {
	return &Client{
		MaxRedirectCount: DefaultMaxRedirectCount,
		Transport:        DefaultTransport(),
		Jar:              specs.NewCookieJar(),
	}
}

// A Client is an HTTP client. Its zero value of [DefaultClient].
//
// A Client is higher-level than a [RoundTripper] (such as [Transport])
// and additionally handles HTTP details such as cookies and
// redirects.
type Client struct {
	// Transport specifies the mechanism by which individual HTTP requests are made.
	// If nil, [DefaultTransport] is used.
	Transport RoundTripper

	// MaxRedirectCount maximum number of redirects
	// before getting an error.
	// if not specified is used [DefaultMaxRedirectCount]
	MaxRedirectCount int

	// Header specifies independent request header and cookies
	//
	// The Header is used to insert headers and cookies
	// into every outbound Request independent of url.
	// The Header is consulted for every redirect that the [Client] follows.
	//
	// If Header is nil, headers and cookies are only sent
	// if they are explicitly set on the [Request].
	Header *specs.Header

	// Jar specifies the cookie jar with dependent to url
	//
	// The Jar is used to insert relevant requested url cookies
	// into every outbound Request and is updated
	// with the cookie values of every inbound Response.
	// The Jar is consulted for every redirect that the [Client] follows.
	//
	// If Jar is nil, cookies are only sent
	// if they are explicitly set on the [Request].
	Jar *specs.CookieJar

	mu sync.RWMutex
}

// Make sends an HTTP request and returns an HTTP response, following
// policy (such as redirects, cookies, auth) as configured on the
// client.
//
// An error is returned if caused by client policy
// or failure to speak HTTP (such as a network connectivity problem).
//
// A non-2xx status code doesn't cause an error.
func (cln *Client) Make(request ClientRequest) (ClientResponse, error) {
	if request == nil {
		panic("plow: nil request pointer")
	}
	return cln.MakeContext(context.Background(), request)
}

// MakeContext version [Client.Make] with [context.Context] cancellation support
func (cln *Client) MakeContext(ctx context.Context, request ClientRequest) (ClientResponse, error) {
	if cln == nil {
		panic("plow: nil client pointer")
	}
	if ctx == nil {
		panic("plow: nil context pointer")
	}
	if request == nil {
		panic("plow: nil request pointer")
	}

	url := request.Url()
	if url.Scheme == "" {
		url.Scheme = "https"
	}

	if !(url.Scheme == "http" || url.Scheme == "https") {
		return nil, fmt.Errorf("invalid request url '%s' scheme", url.Scheme)
	}

	if url.Host == "" {
		return nil, fmt.Errorf("invalid request url '%s' host", url.Host)
	}

	method := request.Method()
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid request method '%s'", method)
	}

	header := request.Header()
	if header == nil {
		panic("plow: nil request.header pointer")
	}

	header = header.Clone()

	maxRedirectCount := DefaultMaxRedirectCount
	if cln.MaxRedirectCount > 0 {
		maxRedirectCount = cln.MaxRedirectCount
	}

	if cln.Jar != nil {
		cln.mu.RLock()
		for cookie := range cln.Jar.Cookies(url.Host) {
			if !header.HasCookie(cookie.Name) {
				header.SetCookie(cookie)
			}
		}
		cln.mu.RUnlock()
	}

	if cln.Header != nil {
		cln.mu.RLock()
		for name, value := range cln.Header.All() {
			if !header.Has(name) {
				header.Set(name, value)
			}
		}
		for cookie := range cln.Header.Cookies() {
			if !header.HasCookie(cookie.Name) {
				header.SetCookie(cookie)
			}
		}
		cln.mu.RUnlock()
	}

	writer, _ := request.(BodyWriter)
	transport := cln.Transport
	if transport == nil {
		cln.mu.Lock()
		if cln.Transport == nil {
			cln.Transport = DefaultTransport()
		}
		transport = cln.Transport
		cln.mu.Unlock()
	}

	var redirectCount int
	for {
		if err := ctx.Err(); err != nil {
			return nil, catch.CatchCommonErr(err)
		}

		resp, err := transport.RoundTrip(ctx, method, url, header, writer)

		if err == nil {
			err = ctx.Err()
		}
		if err != nil {
			return nil, catch.CatchCommonErr(err)
		}

		if cln.Jar != nil {
			cln.mu.Lock()
			cln.Jar.SetCookiesIter(url.Host, resp.Header().Cookies())
			cln.mu.Unlock()
		}

		code := resp.StatusCode()
		if code.IsRedirect() {
			if redirectCount >= maxRedirectCount {
				return nil, specs.NewOpError("redirect", "too many redirects")
			}
			redirectCount++

			if (code == specs.StatusCodeMovedPermanently ||
				code == specs.StatusCodeSeeOther ||
				code == specs.StatusCodeFound) &&
				(method != specs.HttpMethodGet &&
					method != specs.HttpMethodHead) {
				method = specs.HttpMethodGet
			}

			location := resp.Header().Get("Location")
			if location == "" {
				return nil, specs.NewOpError("redirect", "empty Location header")
			}

			var redirectUrl *specs.Url
			redirectUrl, err = specs.ParseUrl(location)
			if err != nil {
				return nil, specs.NewOpError("redirect", "cannot parse location header url")
			}

			if !(redirectUrl.Scheme == "" || redirectUrl.Scheme == "http" || redirectUrl.Scheme == "https") {
				return nil, specs.NewOpError("redirect", "invalid request url '%s' scheme", url.Scheme)
			}

			redirectUrl.Scheme = url.Scheme
			if redirectUrl.Host == "" {
				redirectUrl.Host = url.Host
				redirectUrl.Port = url.Port
			}
			url = *redirectUrl

			continue
		}

		return resp, nil
	}
}
