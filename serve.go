package plow

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"github.com/oesand/plow/internal/catch"
	"github.com/oesand/plow/internal/encoding"
	"github.com/oesand/plow/internal/parsing"
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"strconv"
	"strings"
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
		panic("plow: nil listener")
	}
	if srv.Handler == nil {
		panic("plow: nil server handler")
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
				if err := recover(); err != nil {
					if errorHandler != nil {
						errorHandler.HandleError(ctx, conn, err)
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
					} else if !catch.IsCommonNetReadError(err) {
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
		if srv.TLSHandshakeTimeout > 0 {
			conn.SetDeadline(time.Now().Add(srv.TLSHandshakeTimeout))
		}
		err = catch.CallWithTimeoutContextErr(ctx, srv.TLSHandshakeTimeout, tlsConn.HandshakeContext)

		if err != nil {
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn underlying net.Conn.
			var re tls.RecordHeaderError
			if errors.As(err, &re) && re.Conn != nil && server_ops.TlsRecordHeaderLikeHTTP(re.RecordHeader) {
				return responseErrDowngradeHTTPS
			}
			return err
		}

		if srv.TLSHandshakeTimeout > 0 {
			conn.SetDeadline(time.Time{})
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
		defer conn.SetReadDeadline(time.Time{})
	}

	if srv.WriteTimeout > 0 {
		defer conn.SetWriteDeadline(time.Time{})
	}

	bufioReader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(bufioReader)

	for i := 0; true; i++ {
		if i > 0 {
			idleTimeout := srv.IdleTimeout
			if idleTimeout <= 0 {
				idleTimeout = srv.ReadTimeout
			}

			if idleTimeout > 0 {
				conn.SetReadDeadline(time.Now().Add(idleTimeout))

				// Wait for the connection to become readable again
				// before trying to read the next request.
				if _, err := bufioReader.Peek(4); err != nil {
					return nil
				}

				conn.SetReadDeadline(time.Time{})
			}
		}

		if srv.ReadTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(srv.ReadTimeout))
		}

		if srv.WriteTimeout > 0 {
			conn.SetWriteDeadline(time.Now().Add(srv.WriteTimeout))
		}

		req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), bufioReader, srv.ReadLineMaxLength, srv.HeadMaxLength)

		if err == nil {
			err = ctx.Err()
		}
		if err != nil {
			if !catch.IsCommonNetReadError(err) {
				return responseErrNotProcessable
			}
			return err
		}

		protoMajor, protoMinor := req.ProtoVersion()
		isHttp11 := protoMajor == 1 && protoMinor == 1
		var wantKeepAlive bool
		if !srv.DisableKeepAlive {
			if isHttp11 {
				wantKeepAlive = !strings.EqualFold(req.Header().Get("Connection"), "close")
			} else {
				wantKeepAlive = strings.EqualFold(req.Header().Get("Connection"), "keep-alive")
			}
		}

		// Expect 100 Continue support
		var expectContinue bool
		if expectHeader := req.Header().Get("Expect"); expectHeader != "" {
			if isHttp11 && strings.EqualFold(expectHeader, "100-continue") {
				expectContinue = true
			} else {
				return responseExpectationFailedError
			}
		}

		var selectedEncoding string
		if acceptEncoding, has := req.Header().TryGet("Accept-Encoding"); has {
			variants := strings.Split(acceptEncoding, ", ")
			for _, variant := range variants {
				if encoding.IsKnownEncoding(variant) {
					selectedEncoding = variant
					break
				}
			}
		}

		var isChunked bool
		if req.Method().IsPostable() {
			var contentLength int64
			isChunked, contentLength, err = parsing.ParseContentLength(req.Header())
			if err != nil {
				if errors.Is(err, parsing.ErrParsing) {
					return responseInvalidContentLength
				}
				if errors.Is(err, specs.ErrUnknownTransferEncoding) {
					return responseUnsupportedTransferEncoding
				}

				return err
			}

			if isChunked || contentLength > 0 {
				if srv.MaxBodySize > 0 {
					if isChunked {
						contentLength = srv.MaxBodySize
					} else if contentLength > srv.MaxBodySize {
						return responseErrBodyTooLarge
					}
				}

				var reader io.Reader = bufioReader
				if isChunked {
					reader = encoding.NewChunkedReader(bufioReader)
				}

				if contentLength > 0 {
					reader = io.LimitReader(reader, contentLength)
				}

				if expectContinue {
					_, err = conn.Write(responseContinueBuf)
					if err != nil {
						return err
					}
				}

				req.BodyReader = reader
			}
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

		if selectedEncoding != "" {
			header.Set("Content-Encoding", selectedEncoding)
		}

		var mustClose bool
		if connHeader := header.Get("Connection"); connHeader != "" {
			if strings.EqualFold(connHeader, "close") || !wantKeepAlive {
				mustClose = true
			}
		} else {
			mustClose = !wantKeepAlive
			if mustClose {
				header.Set("Connection", "close")
			} else if !isHttp11 {
				header.Set("Connection", "keep-alive")
			}
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
			if isChunked {
				header.Set("Transfer-Encoding", "chunked")
			} else if header.Get("Transfer-Encoding") == "chunked" {
				isChunked = true
			} else {
				maxEncodingSize := DefaultMaxEncodingSize
				if srv.MaxEncodingSize > 0 {
					maxEncodingSize = srv.MaxEncodingSize
				}
				contentLength := writable.ContentLength()

				if selectedEncoding != "" && contentLength <= maxEncodingSize {
					var cachedBody bytes.Buffer
					err = srv.writeBody(writable, &cachedBody, false, selectedEncoding)
					if err != nil {
						return err
					}
					encodedContent = cachedBody.Bytes()
					header.Set("Content-Length", strconv.Itoa(len(encodedContent)))
				} else {
					selectedEncoding = ""
					if contentLength > 0 {
						header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
					}
				}
			}
		}

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
			err = srv.writeBody(writable, conn, isChunked, selectedEncoding)
		}

		if err != nil {
			return err
		}

		if err = ctx.Err(); err != nil {
			return err
		} else if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(ctx, conn)
			break
		} else if mustClose {
			break
		} else if req.Method() != specs.HttpMethodHead && writable == nil && code.IsReplyable() {
			break
		}
	}

	return nil
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
