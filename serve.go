package giglet

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"net"
	"runtime"
	"time"
)

var ErrorServerShutdown = specs.NewOpError("server", "shutdown")

func (server *Server) Serve(listener net.Listener) error {
	if listener == nil {
		return validationErr("nil listener")
	} else if server.isShuttingdown.Load() {
		return ErrorServerShutdown
	}

	server.listenerTrack.Add(1)
	defer server.listenerTrack.Done()

	for {
		conn, err := listener.Accept()
		if server.isShuttingdown.Load() {
			if err == nil {
				conn.Close()
			}
			return ErrorServerShutdown
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				time.Sleep(time.Second)
				continue
			}
			return err
		}

		ctx := context.Background()
		if server.ConnHandler != nil {
			ctx = server.ConnHandler(conn, ctx)
			if ctx == nil {
				conn.Close()
				continue
			}
		}
		go server.handle(conn, ctx)
	}
}

var bufioReaderPool utils.BufioReaderPool

func (srv *Server) handle(conn net.Conn, ctx context.Context) {
	if srv.Handler == nil || srv.isShuttingdown.Load() {
		conn.Close()
		return
	}
	handler := srv.Handler

	if tlsConn, ok := conn.(*tls.Conn); ok {
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn's underlying net.Conn.
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
				responseErrDowngradeHTTPS.Write(conn)
				conn.Close()
				return
			}
			if srv.Debug {
				srv.logger().Printf("giglet: tls handshake error from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		proto := tlsConn.ConnectionState().NegotiatedProtocol

		if srv.nextProtos != nil {
			if handler, ok := srv.nextProtos[proto]; ok {
				handler(tlsConn)
				return
			}
		}

		conn = tlsConn
	}

	if srv.isShuttingdown.Load() {
		conn.Close()
		return
	}

	reader := bufioReaderPool.Get(conn)

	defer func() {
		if err := recover(); err != nil && err != ErrorCancelled {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			if srv.Debug {
				srv.logger().Printf("http: panic serving %v: %v\n%s", conn.RemoteAddr(), err, buf)
			}
		}

		conn.Close()
		bufioReaderPool.Put(reader)
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		req, err := server.ReadRequest(ctx, conn, reader, srv.ReadTimeout, srv.ReadLineMaxLength, srv.HeadMaxLength)

		if err != nil {
			if srv.Debug {
				srv.logger().Printf("http: read request error from %s: %v", conn.RemoteAddr(), err)
			}
			if !utils.IsCommonNetReadError(err) {
				var respErr *server.ErrorResponse
				if errors.As(err, &respErr) {
					respErr.Write(conn)
				} else {
					responseErrNotProcessable.Write(conn)
				}
			}
			conn.Close()
			break
		}

		resp := handler(req)
		var header *specs.Header
		var code specs.StatusCode
		var writable BodyWriter
		if resp != nil {
			header = resp.Header()
			code = resp.StatusCode()
			writable, _ = resp.(BodyWriter)
		}
		if header == nil {
			header = &specs.Header{}
		}

		if len(srv.ServerName) > 0 {
			header.Set("Server", srv.ServerName)
		} else if len(DefaultServerName) > 0 {
			header.Set("Server", DefaultServerName)
		}

		header.Set("Date", time.Now().Format(specs.TimeFormat))

		if !code.IsValid() {
			if !req.Method().CanHaveResponseBody() || writable == nil {
				code = specs.StatusCodeNoContent
			} else {
				code = specs.StatusCodeOK
			}
		}

		protoMajor, protoMinor := req.ProtoVersion()
		isHttp11 := protoMajor == 1 && protoMinor == 1
		_, err = writing.WriteResponseHead(conn, isHttp11, code, header)
		if err != nil {
			if srv.Debug {
				srv.logger().Printf("http: send response head to %s error: %v", conn.RemoteAddr(), err)
			}
			break
		}
		if req.Method().CanHaveResponseBody() && writable != nil {
			if srv.WriteTimeout > 0 {
				srv.applyWriteTimeout(conn)
			}

			writable.WriteBody(conn)

			if srv.WriteTimeout > 0 {
				conn.SetWriteDeadline(zeroTime)
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if srv.isShuttingdown.Load() {
			break
		} else if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(conn)
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.HaveBody() {
			break
		}
	}
}
