package plow

import (
	"crypto/tls"
	"errors"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/oesand/plow/specs"
)

// DefaultServer factory for creating [Server]
// with optimal parameters for perfomance and safety
//
// Each call creates a new instance of [Server]
// with provided [Server.Handler] parameter
func DefaultServer(handler Handler) *Server {
	if handler == nil {
		panic("plow: handler must not be nil")
	}
	return &Server{
		Handler:             handler,
		ReadLineMaxLength:   1024,
		HeadMaxLength:       8 * 1024,
		MaxBodySize:         10 << 20, // 10 mb
		IdleTimeout:         20 * time.Second,
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxEncodingSize:     DefaultMaxEncodingSize,
	}
}

// A Server defines parameters for running an HTTP server.
// The zero value for Server is a valid configuration.
type Server struct {

	// Handler defines the function to be called for handling each incoming HTTP request.
	// This is the main handler implementing the [Handler] interface, which processes requests and generates responses.
	// Handler must be safe for concurrent use, as the server may invoke it from multiple goroutines simultaneously.
	//
	// If Handler is nil, the server cannot process requests and will panic during initialization.
	// If Handler implements the [ErrorHandler] interface and [Server.ErrorHandler] is nil,
	// it will be used as the ErrorHandler.
	Handler Handler

	// ErrorHandler defines a function for handling errors that occur during HTTP request processing.
	// This can include read/write errors, protocol errors, internal server errors, and other exceptional situations.
	//
	// ErrorHandler is called with error details and request context, allowing custom logic for logging,
	// returning special HTTP responses, or collecting metrics. Must be safe for concurrent use, as the server
	// may invoke it from multiple goroutines simultaneously.
	//
	// If ErrorHandler is not set, the server may use default error handling
	// or try to use the Handler as an ErrorHandler if it implements the [ErrorHandler] interface.
	ErrorHandler ErrorHandler

	// FilterConn handles all new incoming connections to provide filtering by address
	// Returns true - accept, false - close connection
	FilterConn func(addr net.Addr) bool

	// ServerName for sending in response headers.
	ServerName string

	// TLSHandshakeTimeout specifies the maximum amount of time to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	// TLSConfig optionally provides a TLS configuration
	TLSConfig *tls.Config

	// ReadTimeout is the maximum duration for server the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alive are enabled.
	//
	// If zero, the value of ReadTimeout is used.
	// If negative, or if zero and ReadTimeout
	// is zero or negative, there is no timeout.
	IdleTimeout time.Duration

	// ReadLineMaxLength maximum size in bytes
	// to read lines in the request
	// such as headers and headlines
	//
	// If zero there is no limit
	ReadLineMaxLength int64

	// HeadMaxLength maximum size in bytes
	// to read lines in the request
	// such as headline and headers together
	//
	// If zero there is no limit
	HeadMaxLength int64

	// MaxBodySize maximum size in bytes
	// to read request body size.
	//
	// The server responds ErrTooLarge if this limit is greater than 0
	// and response body is greater than the limit.
	//
	// By default, request body size is unlimited.
	MaxBodySize int64

	// MaxEncodingSize maximum size in bytes
	// of the response body that will be encoded (based on the "Accept-Encoding" header)
	// when transfer by size - "Content-Length", except { "Transfer-Encoding": "chunked" }
	//
	// if not specified, the encoding will be skipped
	MaxEncodingSize int64

	// DisableKeepAlive controls whether HTTP keep-alive are enabled.
	//
	// Only very resource-constrained environments or servers in the process of
	// shutting down should disable them.
	//
	// By default, keep-alive are always enabled.
	DisableKeepAlive bool

	tlsNextProtos map[string]NextProtoHandler

	listenerTrack sync.WaitGroup
	shuttingDown  chan struct{}

	mutex sync.Mutex
	once  sync.Once
}

func (srv *Server) beforeOnce() {
	srv.shuttingDown = make(chan struct{})
}

// TLSHasNextProto checks if a handler function is specified
// for the upgraded protocol during a TLS connection
// when an ALPN protocol upgrade has occurred.
//
// The proto is the protocol name negotiated.
//
// The handle function can be specified using TLSNextProto
func (srv *Server) TLSHasNextProto(proto string) bool {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	var has bool
	if srv.tlsNextProtos != nil {
		_, has = srv.tlsNextProtos[proto]
	}

	return has
}

// TLSNextProto specifies a function to take over
// ownership of the provided TLS connection when an ALPN
// protocol upgrade has occurred.
//
// The proto is the protocol name negotiated.
//
// [NextProtoHandler] argument should be used to
// handle HTTP requests. The connection is automatically closed
// when the function returns.
//
// HTTP/2 support is not enabled automatically.
func (srv *Server) TLSNextProto(proto string, handler NextProtoHandler) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	if srv.tlsNextProtos == nil {
		srv.tlsNextProtos = map[string]NextProtoHandler{}
	}
	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}
	if !slices.Contains(srv.TLSConfig.NextProtos, proto) {
		srv.TLSConfig.NextProtos = append(srv.TLSConfig.NextProtos, proto)
	}
	srv.tlsNextProtos[proto] = handler
}

// ListenAndServe listens on the TCP network address and then
// calls [Server.Serve] to handle requests on incoming connections.
//
// If addr is blank, ":http" is used.
//
// ListenAndServe always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) ListenAndServe(addr string) error {
	if srv.shuttingDown != nil {
		select {
		case <-srv.shuttingDown:
			return specs.ErrClosed
		default:
		}
	}
	if srv.IsShutdown() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return srv.Serve(lst)
}

// ListenAndServeTLS listens on the TCP network address and then
// calls [Server.ServeTLS] to handle requests on incoming connections.
//
// Filenames containing a certificate and matching private key for the
// server must be provided.
//
// If addr is blank, ":http" is used.
//
// ListenAndServeTLS always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	if srv.IsShutdown() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return srv.ServeTLS(lst, certFile, keyFile)
}

// ListenAndServeTLSRaw listens on the TCP network address and then
// calls [Server.ServeTLSRaw] to handle requests on incoming connections.
//
// Certificate and matching private key for the server must be provided.
//
// If addr is blank, ":http" is used.
//
// ListenAndServeTLSRaw always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) ListenAndServeTLSRaw(addr string, cert tls.Certificate) error {
	if srv.IsShutdown() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return srv.serveTLSRaw(lst, &cert)
}

// ServeTLS accepts incoming connections on the [net.Listener], creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling [Server.Handler] to reply to them.
//
// Filenames containing a certificate and matching private key for the
// server must be provided.
//
// ServeTLS always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) ServeTLS(lst net.Listener, certFile, keyFile string) error {
	if srv.IsShutdown() {
		return specs.ErrClosed
	} else if len(certFile) == 0 || len(keyFile) == 0 {
		return errors.New("plow: unknown certificate source")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return srv.serveTLSRaw(lst, &cert)
}

// ServeTLSRaw accepts incoming connections on the [net.Listener], creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling [Server.Handler] to reply to them.
//
// Certificate and matching private key for the server must be provided.
//
// ServeTLSRaw always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) ServeTLSRaw(lst net.Listener, cert tls.Certificate) error {
	return srv.serveTLSRaw(lst, &cert)
}

func (srv *Server) serveTLSRaw(lst net.Listener, cert *tls.Certificate) error {
	if srv.IsShutdown() {
		return specs.ErrClosed
	}

	var config *tls.Config
	if srv.TLSConfig != nil {
		config = srv.TLSConfig.Clone()
	} else {
		config = &tls.Config{}
	}

	if !slices.Contains(config.NextProtos, httpV1NextProtoTLS) {
		config.NextProtos = append(config.NextProtos, httpV1NextProtoTLS)
	}

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert {
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0] = *cert
	}

	listener := tls.NewListener(lst, config)
	return srv.Serve(listener)
}

// IsShutdown checks if the server is shutting down
// or is already shut down after calling [Server.Shutdown]
func (srv *Server) IsShutdown() bool {
	if srv.shuttingDown != nil {
		select {
		case <-srv.shuttingDown:
			return true
		default:
		}
	}
	return false
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections. Shutdown works by first closing all open
// listeners, then closing all idle connections, and then waiting
// indefinitely for connections to return to idle and then shutdown.
// If the provided context expires before the shutdown is complete,
// Shutdown returns the context's error, otherwise it returns any
// error returned from closing the [Server]'s underlying Listener(s).
//
// When Shutdown is called, [Server.Serve], [Server.ListenAndServe], etc.
// immediately return [ErrServerClosed]. Make sure the
// program doesn't exit and waits instead for Shutdown to return.
//
// Shutdown does not attempt to close nor wait for hijacked
// connections such as WebSockets. The caller of Shutdown should
// separately notify such long-lived connections of shutdown and wait
// for them to close.
//
// Once Shutdown has been called on a server, it may not be reused;
// future calls to methods such as Serve will return ErrServerClosed.
func (srv *Server) Shutdown() {
	if srv.shuttingDown == nil {
		srv.shuttingDown = make(chan struct{})
	}
	close(srv.shuttingDown)

	srv.listenerTrack.Wait()
}
