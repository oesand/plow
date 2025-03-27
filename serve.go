package giglet

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"runtime"
	"time"
)

var ErrorServerShutdown = &specs.GigletError{
	Op:  specs.GigletOp("server"),
	Err: errors.New("shutdown"),
}

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

var bufioReaderPool internal.BufioReaderPool

func (server *Server) handle(conn net.Conn, ctx context.Context) {
	if server.Handler == nil || server.isShuttingdown.Load() {
		conn.Close()
		return
	}
	handler := server.Handler

	if tlsConn, ok := conn.(*tls.Conn); ok {
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn's underlying net.Conn.
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
				re.Conn.Write(responseDowngradeHTTPS)
				re.Conn.Close()
				return
			}
			if server.Debug {
				server.logger().Printf("giglet: tls handshake error from %s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		proto := tlsConn.ConnectionState().NegotiatedProtocol

		if server.nextProtos != nil {
			if handler, ok := server.nextProtos[proto]; ok {
				handler(tlsConn)
				return
			}
		}

		conn = tlsConn
	}

	if server.isShuttingdown.Load() {
		conn.Close()
		return
	}

	reader := bufioReaderPool.Get(conn)

	defer func() {
		if err := recover(); err != nil && err != ErrorCancelled {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			if server.Debug {
				server.logger().Printf("http: panic serving %v: %v\n%s", conn.RemoteAddr(), err, buf)
			}
		}

		conn.Close()
		bufioReaderPool.Put(reader)
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		if server.ContentMaxSizeBytes > 0 {
			reader.Reset(io.LimitReader(conn, server.ContentMaxSizeBytes))
		} else if DefaultContentMaxSizeBytes > 0 {
			reader.Reset(io.LimitReader(conn, DefaultContentMaxSizeBytes))
		}

		server.applyReadTimeout(conn)
		req, err := readRequest(ctx, reader)
		conn.SetReadDeadline(zeroTime)

		req.server = server
		req.conn = conn
		req.context = ctx

		if err != nil {
			if server.Debug {
				server.logger().Printf("http: read request error from %s: %v", conn.RemoteAddr(), err)
			}
			switch {

			case specs.MatchError(err, ErrorReadingUnsupportedEncoding):
				conn.Write(responseUnsupportedEncoding)

			default:
				if !internal.IsCommonNetReadError(err) {
					if serr, ok := err.(*statusErrorResponse); ok {
						serr.Write(conn)
					} else {
						conn.Write(responseNotProcessableError)
					}
				}

			}
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

		if len(server.ServerName) > 0 {
			header.Set("Server", server.ServerName)
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

		_, err = writeResponseHead(conn, req.ProtoAtLeast(1, 1), code, header)
		if err != nil {
			if server.Debug {
				server.logger().Printf("http: send response head to %s error: %v", conn.RemoteAddr(), err)
			}
			break
		}
		if req.method.CanHaveResponseBody() && writable != nil {
			if server.WriteTimeout > 0 {
				server.applyWriteTimeout(conn)
			}

			writable.WriteBody(conn)

			if server.WriteTimeout > 0 {
				conn.SetWriteDeadline(zeroTime)
			}
		}

		if req.cachedMultipart != nil {
			req.cachedMultipart.RemoveAll()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		if server.isShuttingdown.Load() {
			break
		} else if req.hijacker != nil {
			req.hijacker(conn)
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.HaveBody() {
			break
		}
	}
}
