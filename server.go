package giglet

import (
	"crypto/tls"
	"log"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	// Handler to invoke
	Handler Handler

	Logger *log.Logger

	// Handler for new incoming connections to provide filtering by address
	// Returns context.Context - accept, nil - close connection
	ConnHandler ConnHandler

	// Debug flag to allow show system messages
	Debug bool

	// Server name for sending in response headers.
	ServerName string

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. A zero or negative value means
	// there will be no timeout.
	WriteTimeout time.Duration

	// TLSConfig optionally provides a TLS configuration
	TLSConfig *tls.Config

	// ContentMaxSizeBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line and the request body.
	// If zero, DefaultContentMaxSizeBytes is used.
	ContentMaxSizeBytes int64

	nextProtos     map[string]NextProtoHandler
	isShuttingdown atomic.Bool
	listenerTrack  sync.WaitGroup

	mutex      sync.Mutex
	onShutdown []EventHandler
}

func (server *Server) logger() *log.Logger {
	if server.Logger != nil {
		return server.Logger
	}
	return log.Default()
}

func (server *Server) applyReadTimeout(conn net.Conn) {
	if server.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(server.ReadTimeout))
	}
}

func (server *Server) applyWriteTimeout(conn net.Conn) {
	if server.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(server.WriteTimeout))
	}
}

func (server *Server) HasNextProto(proto string) bool {
	server.mutex.Lock()
	_, has := server.nextProtos[proto]
	server.mutex.Unlock()

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
		return ErrorServerShutdown
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
		return ErrorServerShutdown
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
		return ErrorServerShutdown
	} else if addr == "" {
		addr = ":http"
	}
	lst, err := net.Listen("tcp4", addr)
	if err != nil {
		return err
	}
	return server.ServeTLSRaw(lst, cert)
}

func (srv *Server) ServeTLS(lst net.Listener, certFile, keyFile string) error {
	if len(certFile) == 0 || len(keyFile) == 0 {
		return validationErr("unknown certificate source")
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return srv.ServeTLSRaw(lst, cert)
}

func (server *Server) ServeTLSRaw(lst net.Listener, cert tls.Certificate) error {
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
		config.Certificates[0] = cert
	}

	listener := tls.NewListener(lst, config)
	return server.Serve(listener)
}

func (server *Server) OnShutdown(handler EventHandler) {
	if server.isShuttingdown.Load() {
		return
	}
	server.mutex.Lock()
	server.onShutdown = append(server.onShutdown, handler)
	server.mutex.Unlock()
}

func (server *Server) Shutdown() {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	for _, handle := range server.onShutdown {
		go handle()
	}

	server.isShuttingdown.Store(true)
	server.listenerTrack.Wait()
}
