package giglet

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/encoding"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"
)

func (srv *Server) Serve(listener net.Listener) error {
	if listener == nil {
		panic("nil listener")
	}
	if srv.Handler == nil {
		panic("nil server handler")
	}
	srv.once.Do(srv.beforeOnce)

	if srv.IsShutdown() {
		return specs.ErrClosed
	}

	handler := srv.Handler

	srv.listenerTrack.Add(1)
	defer srv.listenerTrack.Done()

	var attemptDelay time.Duration
	var connTrack sync.WaitGroup
	var err error

	ctx, cancelCtx := context.WithCancel(context.Background())
	for {
		var conn net.Conn
		conn, err = srv.accept(ctx, listener)

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
		if srv.FilterConn != nil {
			if allow := srv.FilterConn(conn.RemoteAddr()); !allow {
				conn.Close()
				continue
			}
		}

		connTrack.Add(1)
		go func() {
			srv.handle(ctx, conn, handler)
			connTrack.Done()
		}()
	}

	cancelCtx()
	listener.Close()
	connTrack.Wait()

	return err
}

func (srv *Server) accept(ctx context.Context, listener net.Listener) (net.Conn, error) {
	connRes := make(chan catch.ResultErrPair[net.Conn], 1)

	go func() {
		conn, err := listener.Accept()
		connRes <- catch.ResultErrPair[net.Conn]{conn, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-srv.shuttingDown:
		return nil, specs.ErrClosed
	case res := <-connRes:
		return res.Res, res.Err
	}
}

func (srv *Server) handle(ctx context.Context, conn net.Conn, handler Handler) {
	if srv.catchCancelled(ctx) {
		conn.Close()
		return
	}

	if tlsConn, ok := conn.(*tls.Conn); ok {
		err := catch.CallWithTimeoutContextErr(ctx, srv.TLSHandshakeTimeout, tlsConn.HandshakeContext)

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
		req, err := server.ReadRequest(ctx, conn.RemoteAddr(), headerReader, srv.ReadLineMaxLength, srv.HeadMaxLength)

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

			if req.Chunked {
				if srv.MaxBodySize > 0 {
					reader = io.LimitReader(reader, srv.MaxBodySize)
				}
				reader = httputil.NewChunkedReader(reader)
			} else {
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

		resp := handler.Handle(ctx, req)
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

		if req.SelectedEncoding != "" {
			header.Set("Content-Encoding", req.SelectedEncoding)
		}

		if !code.IsValid() {
			if !req.Method().IsReplyable() || writable == nil {
				code = specs.StatusCodeNoContent
			} else {
				code = specs.StatusCodeOK
			}
		}

		var encodedContent []byte
		shouldResponseBody := req.Method().IsReplyable() && code.IsReplyable() && writable != nil
		if shouldResponseBody {
			if req.Chunked {
				header.Set("Transfer-Encoding", "chunked")
			} else if header.Get("Transfer-Encoding") == "chunked" {
				req.Chunked = true
			} else {
				maxEncodingSize := DefaultMaxEncodingSize
				if srv.MaxEncodingSize > 0 {
					maxEncodingSize = srv.MaxEncodingSize
				}
				contentLength := writable.ContentLength()

				if req.SelectedEncoding != "" && contentLength <= maxEncodingSize {
					var cachedBody bytes.Buffer
					err = srv.writeBody(writable, &cachedBody, false, req.SelectedEncoding)
					if err != nil {
						if srv.Debug {
							srv.logger().Printf("giglet: fail to cache encoded response '%s': %v", conn.RemoteAddr(), err)
						}
					}
					encodedContent = cachedBody.Bytes()
					header.Set("Content-Length", strconv.Itoa(len(encodedContent)))
				} else {
					req.SelectedEncoding = ""
					if contentLength > 0 {
						header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
					}
				}
			}
		}

		protoMajor, protoMinor := req.ProtoVersion()
		isHttp11 := protoMajor == 1 && protoMinor == 1

		if srv.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(srv.WriteTimeout))
		}
		_, err = server.WriteResponseHead(conn, isHttp11, code, header)

		if err != nil {
			if srv.Debug {
				srv.logger().Printf("giglet: error to send head to '%s': %v", conn.RemoteAddr(), err)
			}
			break
		}

		if srv.catchCancelled(ctx) {
			break
		}

		if encodedContent != nil {
			if len(encodedContent) > 0 {
				_, err = conn.Write(encodedContent)
			}
		} else if shouldResponseBody {
			err = srv.writeBody(writable, conn, req.Chunked, req.SelectedEncoding)
		}

		if err != nil {
			if srv.Debug {
				srv.logger().Printf("giglet: error to send body to '%s': %v", conn.RemoteAddr(), err)
			}
			break
		}

		if srv.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Time{})
		}

		if srv.catchCancelled(ctx) {
			break
		} else if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(ctx, conn)
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.IsReplyable() {
			break
		}
	}
}

func (srv *Server) catchCancelled(ctx context.Context) bool {
	return ctx.Err() != nil
}

func (srv *Server) writeBody(writable BodyWriter, writer io.Writer, chunked bool, contentEncoding string) error {
	if chunked {
		chw := encoding.NewChunkedWriter(writer)
		defer chw.Close()
		writer = chw
	}

	if contentEncoding != "" {
		encodingWriter, err := encoding.NewWriter(contentEncoding, writer)
		if err != nil {
			return err
		}
		defer encodingWriter.Close()
		writer = encodingWriter
	}

	return writable.WriteBody(writer)
}
