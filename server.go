package giglet

import (
	"crypto/tls"
	"github.com/oesand/giglet/specs"
	"log"
	"net"
	"slices"
	"sync"
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

	listenerTrack sync.WaitGroup
	shuttingDown  chan struct{}

	mutex sync.Mutex
	once  sync.Once
}

func (srv *Server) beforeOnce() {
	srv.shuttingDown = make(chan struct{})
}

func (srv *Server) logger() *log.Logger {
	if srv.Logger != nil {
		return srv.Logger
	}
	return log.Default()
}

func (srv *Server) HasNextProto(proto string) bool {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	_, has := srv.nextProtos[proto]
	return has
}

func (srv *Server) NextProto(proto string, handler NextProtoHandler) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	if srv.nextProtos == nil {
		srv.nextProtos = map[string]NextProtoHandler{}
	}
	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}
	if !slices.Contains(srv.TLSConfig.NextProtos, proto) {
		srv.TLSConfig.NextProtos = append(srv.TLSConfig.NextProtos, proto)
	}
	srv.nextProtos[proto] = handler
}

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

func (srv *Server) ServeTLS(lst net.Listener, certFile, keyFile string) error {
	if srv.IsShutdown() {
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

func (srv *Server) Shutdown() {
	if srv.shuttingDown == nil {
		srv.shuttingDown = make(chan struct{})
	}
	close(srv.shuttingDown)

	srv.listenerTrack.Wait()
}
