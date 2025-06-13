package giglet

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"
)

func (server *Server) Serve(listener net.Listener) error {
	if listener == nil {
		panic("nil listener")
	}
	if server.Handler == nil {
		panic("nil server handler")
	}
	if server.isShuttingdown.Load() {
		return specs.ErrClosed
	}

	handler := server.Handler

	server.listenerTrack.Add(1)
	defer server.listenerTrack.Done()

	var attemptDelay time.Duration
	var connTrack sync.WaitGroup
	var err error

	ctx, cancelCtx := context.WithCancel(context.Background())
	for {
		var conn net.Conn
		conn, err = listener.Accept()
		if server.isShuttingdown.Load() {
			if err == nil && conn != nil {
				conn.Close()
			}
			err = specs.ErrClosed
			break
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if attemptDelay == 0 {
					attemptDelay = 5 * time.Millisecond
				} else if maxDelay := 1 * time.Second; attemptDelay >= maxDelay {
					attemptDelay = maxDelay
				} else {
					attemptDelay *= 2
				}

				time.Sleep(attemptDelay)
				continue
			}
			break
		}

		attemptDelay = 0
		connTrack.Add(1)

		if server.FilterConn != nil {
			if allow := server.FilterConn(conn.RemoteAddr()); !allow {
				connTrack.Done()
				conn.Close()
				continue
			}
		}

		go func() {
			server.handle(ctx, conn, handler)
			connTrack.Done()
		}()
	}

	cancelCtx()
	connTrack.Wait()
	listener.Close()

	return err
}

func (srv *Server) catchCancelled(ctx context.Context) bool {
	return ctx.Err() != nil
}

func (srv *Server) handle(ctx context.Context, conn net.Conn, handler Handler) {
	if srv.catchCancelled(ctx) {
		conn.Close()
		return
	}

	if tlsConn, ok := conn.(*tls.Conn); ok {
		_, err := catch.CallWithTimeoutContext(ctx, srv.TLSHandshakeTimeout, func(ctx context.Context) (struct{}, error) {
			err := tlsConn.HandshakeContext(ctx)
			return struct{}{}, err
		})

		if err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn underlying net.Conn.
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
				responseErrDowngradeHTTPS.WriteTo(conn)
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
	}

	defer func() {
		if err := recover(); err != nil && srv.Debug {
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
		req, err := server.ReadRequest(ctx, conn, headerReader, srv.ReadLineMaxLength, srv.HeadMaxLength)

		if err != nil {
			stream.DefaultBufioReaderPool.Put(headerReader)
			if srv.Debug {
				srv.logger().Printf("giglet: read request error from '%s': %v", conn.RemoteAddr(), err)
			}
			if !catch.IsCommonNetReadError(err) {
				var respErr *server.ErrorResponse
				if errors.As(err, &respErr) {
					respErr.WriteTo(conn)
				} else {
					responseErrNotProcessable.WriteTo(conn)
				}
			}
			break
		}

		if srv.catchCancelled(ctx) {
			stream.DefaultBufioReaderPool.Put(headerReader)
			break
		}

		if req.Method().IsPostable() {
			extraBuffered, _ := headerReader.Peek(headerReader.Buffered())

			reader := io.MultiReader(bytes.NewReader(extraBuffered), conn)

			////
			if encoding, has := req.Header().TryGet("Transfer-Encoding"); has {
				switch encoding {
				case "chunked":
					if srv.MaxBodySize > 0 {
						reader = io.LimitReader(reader, srv.MaxBodySize)
					}
					reader = httputil.NewChunkedReader(reader)
				}
			} else if raw, has := req.Header().TryGet("Content-Length"); has && len(raw) > 0 {
				if contentLength, err := strconv.ParseInt(raw, 10, 64); err != nil {
					reader = io.LimitReader(reader, contentLength)
				}
			}

			switch req.Header().Get("Transfer-Encoding") {
			case "chunked":
				if srv.MaxBodySize > 0 {
					reader = io.LimitReader(reader, srv.MaxBodySize)
				}
				reader = httputil.NewChunkedReader(reader)
			default:
				var contentLength int64
				if cl := req.Header().Get("Content-Length"); cl != "" {
					contentLength, err = strconv.ParseInt(cl, 10, 64)
					if err == nil {
						if contentLength <= 0 {
							reader = nil
						} else if srv.MaxBodySize > 0 && contentLength >= srv.MaxBodySize {
							responseErrBodyTooLarge.WriteTo(conn)
							return
						}
					}
				}

				if reader != nil {
					if contentLength <= 0 && srv.MaxBodySize > 0 {
						contentLength = srv.MaxBodySize
					}

					if contentLength > 0 {
						reader = io.LimitReader(reader, contentLength)
					}
				}
			}

			req.BodyReader = reader
		}

		stream.DefaultBufioReaderPool.Put(headerReader)

		if srv.catchCancelled(ctx) {
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
			if !req.Method().IsReplyable() || writable == nil {
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

		if srv.catchCancelled(ctx) {
			break
		}

		if req.Method().IsReplyable() && code.IsReplyable() && writable != nil {
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

		if srv.catchCancelled(ctx) {
			break
		} else if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(conn)
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.IsReplyable() {
			break
		}
	}
}
