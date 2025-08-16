package giglet

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/encoding"
	"github.com/oesand/giglet/internal/server_ops"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"
)

// Serve accepts incoming connections on the [net.Listener], creating a
// new service goroutine for each. The service goroutines read requests and
// then call [Server.Handler] to reply to them.
//
// HTTP/2 not supported
//
// Serve always returns a non-nil error.
// After [Server.Shutdown], the returned error is [specs.ErrClosed].
func (srv *Server) Serve(listener net.Listener) error {
	if listener == nil {
		panic("nil listener")
	}
	if srv.Handler == nil {
		panic("nil server handler")
	}
	if srv.IsShutdown() {
		return specs.ErrClosed
	}

	srv.once.Do(srv.beforeOnce)

	handler := srv.Handler
	errorHandler := srv.ErrorHandler
	if errorHandler == nil {
		errorHandler, _ = handler.(ErrorHandler)
	}

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
		go func(conn net.Conn) {
			defer connTrack.Done()
			defer func() {
				if r := recover(); r != nil {
					if errorHandler != nil {
						errorHandler.HandleError(ctx, conn, r)
					} else {
						conn.SetDeadline(time.Now().Add(time.Second))
						responseInternalServerError.WriteTo(conn)
					}
				}
				conn.Close()
			}()

			if err := srv.handle(ctx, conn, handler); err != nil {
				if errorHandler != nil {
					errorHandler.HandleError(ctx, conn, err)
				} else {
					conn.SetDeadline(time.Now().Add(time.Second))
					var respErr *server_ops.ErrorResponse
					if errors.As(err, &respErr) {
						respErr.WriteTo(conn)
					} else {
						responseInternalServerError.WriteTo(conn)
					}
				}
			}
		}(conn)
	}

	cancelCtx()
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

func (srv *Server) handle(ctx context.Context, conn net.Conn, handler Handler) error {
	var err error
	if err = ctx.Err(); err != nil {
		return err
	}

	if tlsConn, ok := conn.(*tls.Conn); ok {
		err = catch.CallWithTimeoutContextErr(ctx, srv.TLSHandshakeTimeout, tlsConn.HandshakeContext)

		if err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn underlying net.Conn.
			if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil {
				return responseErrDowngradeHTTPS
			}
			return err
		}

		proto := tlsConn.ConnectionState().NegotiatedProtocol

		if srv.tlsNextProtos != nil {
			if handler, ok := srv.tlsNextProtos[proto]; ok {
				handler(tlsConn)
				return nil
			}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if srv.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(srv.ReadTimeout))
	}

	if srv.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(srv.WriteTimeout))
	}

	for {
		req, extraBuffered, err := srv.readHeader(ctx, conn)

		if err != nil {
			if !catch.IsCommonNetReadError(err) {
				return responseErrNotProcessable
			}
			return err
		}

		if err = ctx.Err(); err != nil {
			return err
		}

		if req.Method().IsPostable() {
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
							return responseErrBodyTooLarge
						}
					} else {
						return errors.New("invalid Content-Length header: " + cl)
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

		if err = ctx.Err(); err != nil {
			return err
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
			header = specs.NewHeader()
		}

		if srv.ServerName != "" {
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
		mustResponseBody := req.Method().IsReplyable() && code.IsReplyable() && writable != nil
		if mustResponseBody {
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
						return err
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

		_, err = server_ops.WriteResponseHead(conn, isHttp11, code, header)

		if err != nil {
			return err
		}

		if err = ctx.Err(); err != nil {
			return err
		}

		if encodedContent != nil {
			if len(encodedContent) > 0 {
				_, err = conn.Write(encodedContent)
			}
		} else if mustResponseBody {
			err = srv.writeBody(writable, conn, req.Chunked, req.SelectedEncoding)
		}

		if err != nil {
			return err
		}

		if err = ctx.Err(); err != nil {
			return err
		} else if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(ctx, conn)
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.IsReplyable() {
			break
		}
	}

	if srv.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Time{})
	}

	if srv.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Time{})
	}

	return nil
}

func (srv *Server) readHeader(ctx context.Context, conn net.Conn) (*server_ops.HttpRequest, []byte, error) {
	headerReader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(headerReader)

	req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), headerReader, srv.ReadLineMaxLength, srv.HeadMaxLength)

	if err != nil {
		return nil, nil, err
	}

	extraBuffered, _ := headerReader.Peek(headerReader.Buffered())
	return req, extraBuffered, nil
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
