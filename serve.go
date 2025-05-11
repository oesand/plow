package giglet

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"time"
)

func (server *Server) Serve(listener net.Listener) error {
	if listener == nil {
		return validationErr("nil listener")
	} else if server.isShuttingdown.Load() {
		return specs.ErrCancelled
	}

	server.listenerTrack.Add(1)
	defer server.listenerTrack.Done()

	for {
		conn, err := listener.Accept()
		if server.isShuttingdown.Load() {
			if err == nil {
				conn.Close()
			}
			return specs.ErrCancelled
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
				srv.logger().Printf("giglet: tls handshake error from '%s': %v", conn.RemoteAddr(), err)
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

	defer func() {
		if err := recover(); err != nil && err != specs.ErrCancelled && srv.Debug {
			srv.logger().Printf("giglet: panic serving from '%s': %v", conn.RemoteAddr(), err)
		}

		conn.Close()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		headerReader := stream.DefaultBufioReaderPool.Get(conn)

		if srv.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(srv.ReadTimeout))
		}
		req, err := server.ReadRequest(ctx, conn, headerReader, srv.ReadTimeout, srv.ReadLineMaxLength, srv.HeadMaxLength)

		if err != nil {
			stream.DefaultBufioReaderPool.Put(headerReader)
			if srv.Debug {
				srv.logger().Printf("giglet: read request error from '%s': %v", conn.RemoteAddr(), err)
			}
			if !stream.IsCommonNetReadError(err) {
				var respErr *server.ErrorResponse
				if errors.As(err, &respErr) {
					respErr.Write(conn)
				} else {
					responseErrNotProcessable.Write(conn)
				}
			}
			break
		}

		if req.Method().IsPostable() {
			extraBuffered, _ := headerReader.Peek(headerReader.Buffered())

			reader := io.MultiReader(bytes.NewReader(extraBuffered), conn)

			if encoding, has := req.Header().TryGet("Transfer-Encoding"); has {
				switch encoding {
				case "chunked":
					reader = httputil.NewChunkedReader(reader)
				}
			} else if raw, has := req.Header().TryGet("Content-Length"); has && len(raw) > 0 {
				if contentLength, err := strconv.ParseInt(raw, 10, 64); err != nil {
					reader = io.LimitReader(reader, contentLength)
				}
			}

			req.BodyReader = reader
		}

		stream.DefaultBufioReaderPool.Put(headerReader)

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

		if srv.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Time{})
		}

		if len(srv.ServerName) > 0 {
			header.Set("Server", srv.ServerName)
		} else {
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

		if srv.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(srv.WriteTimeout))
		}
		_, err = writing.WriteResponseHead(conn, isHttp11, code, header)

		if err != nil {
			if srv.Debug {
				srv.logger().Printf("giglet: error to send head to '%s': %v", conn.RemoteAddr(), err)
			}
			break
		}

		if req.Method().CanHaveResponseBody() && writable != nil {
			var writer io.Writer = conn
			var encodingCloser io.Closer

			switch req.SelectedEncoding {
			case specs.GzipContentEncoding:
				gzw := gzip.NewWriter(writer)
				writer = gzw
				encodingCloser = gzw
			}

			if req.SelectedEncoding != specs.UnknownContentEncoding {
				resp.Header().Set("Content-Encoding", string(req.SelectedEncoding))
			}

			err = writable.WriteBody(writer)

			if err == nil && encodingCloser != nil {
				err = encodingCloser.Close()
			}
			if err != nil {
				if srv.Debug {
					srv.logger().Printf("giglet: error to send body to '%s': %v", conn.RemoteAddr(), err)
				}
				break
			}

		}

		if srv.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Time{})
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
