package giglet

import (
	"crypto/tls"
	"github.com/oesand/giglet/specs"
	"log"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

func DefaultServer(handler Handler) *Server {
	if handler == nil {
		panic("handler must not be nil")
	}
	return &Server{
		Handler:             handler,
		ReadLineMaxLength:   1024,
		HeadMaxLength:       8 * 1024,
		MaxBodySize:         10 << 20, // 10 mb
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        10 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxEncodingSize:     DefaultMaxEncodingSize,
	}
}

type Server struct {
	// Handler to invoke
	Handler Handler

	Logger *log.Logger

	// FilterConn handles all new incoming connections to provide filtering by address
	// Returns true - accept, false - close connection
	FilterConn func(addr net.Addr) bool

	// Debug flag to allow show system messages
	Debug bool

	// ServerName for sending in response headers.
	ServerName string

	// ReadTimeout is the maximum duration for server the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

	// TLSHandshakeTimeout specifies the maximum amount of time to
	// wait for a TLS handshake. Zero means no timeout.
	TLSHandshakeTimeout time.Duration

	// TLSConfig optionally provides a TLS configuration
	TLSConfig *tls.Config

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

	nextProtos map[string]NextProtoHandler

	listenerTrack  sync.WaitGroup
	isShuttingdown atomic.Bool

	mutex sync.Mutex
}

func (server *Server) logger() *log.Logger {
	if server.Logger != nil {
		return server.Logger
	}
	return log.Default()
}

func (server *Server) HasNextProto(proto string) bool {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	_, has := server.nextProtos[proto]
	return has
}

func (server *Server) NextProto(proto string, handler NextProtoHandler) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if server.nextProtos == nil {
		server.nextProtos = map[string]NextProtoHandler{}
	}
	if server.TLSConfig == nil {
		server.TLSConfig = &tls.Config{}
	}
	if !slices.Contains(server.TLSConfig.NextProtos, proto) {
		server.TLSConfig.NextProtos = append(server.TLSConfig.NextProtos, proto)
	}
	server.nextProtos[proto] = handler
}

func (server *Server) ListenAndServe(addr string) error {
	if server.isShuttingdown.Load() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return server.Serve(lst)
}

func (server *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	if server.isShuttingdown.Load() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return server.ServeTLS(lst, certFile, keyFile)
}

func (server *Server) ListenAndServeTLSRaw(addr string, cert tls.Certificate) error {
	if server.isShuttingdown.Load() {
		return specs.ErrClosed
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return server.serveTLSRaw(lst, &cert)
}

func (srv *Server) ServeTLS(lst net.Listener, certFile, keyFile string) error {
	if srv.isShuttingdown.Load() {
		return specs.ErrClosed
	} else if len(certFile) == 0 || len(keyFile) == 0 {
		panic("unknown certificate source")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return srv.serveTLSRaw(lst, &cert)
}

func (server *Server) ServeTLSRaw(lst net.Listener, cert tls.Certificate) error {
	return server.serveTLSRaw(lst, &cert)
}

func (server *Server) serveTLSRaw(lst net.Listener, cert *tls.Certificate) error {
	if server.isShuttingdown.Load() {
		return specs.ErrClosed
	}

	var config *tls.Config
	if server.TLSConfig != nil {
		config = server.TLSConfig.Clone()
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
	return server.Serve(listener)
}

func (server *Server) Shutdown() {
	server.isShuttingdown.Store(true)
	server.listenerTrack.Wait()
}
