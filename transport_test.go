package plow

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"github.com/armon/go-socks5"
	"github.com/oesand/plow/internal/encoding"
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestTransport_GetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-hello-world", "xyz-123")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	resp, err := DefaultTransport().RoundTrip(
		context.Background(), specs.HttpMethodGet, *specs.MustParseUrl(server.URL), specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Hello-World") != "xyz-123" ||
		resp.Header().Get("Content-Encoding") != "" ||
		resp.Header().Get("Content-Type") != "application/json" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, []byte("OK"))
}

func TestTransport_PostRequest(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Hello-World") != "xyz-123" ||
			r.Header.Get("Content-Length") != strconv.Itoa(len(requestBody)) ||
			r.Header.Get("Content-Type") != specs.ContentTypePlain ||
			r.Header.Get("X-Type") != "json" {
			t.Errorf("not found expected headers: %+v", r.Header)
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	req := BufferRequest(specs.HttpMethodPost, specs.MustParseUrl(server.URL), specs.ContentTypePlain, requestBody)
	req.Header().Set("x-type", "json")
	req.Header().Set("x-hello-world", "xyz-123")

	resp, err := DefaultTransport().RoundTrip(
		context.Background(), req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

// Test chunked transfer

func TestTransport_ChunkedTransferEncoding(t *testing.T) {
	testContent := []byte("Chunked\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		header := specs.NewHeader()
		header.Set("Transfer-Encoding", "chunked")

		var cacheBuf bytes.Buffer
		cw := encoding.NewChunkedWriter(&cacheBuf)
		cw.Write(testContent)
		cw.Close()
		return specs.StatusCodeOK, header, cacheBuf.Bytes()
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Transfer-Encoding") != "chunked" {
		t.Errorf("expected chunked, got %s", resp.Header().Get("Transfer-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

func TestTransport_PostChunkedTransferEncodingRequestHttpTest(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Hello-World") != "xyz-123" ||
			r.Header.Get("Content-Length") != "" {
			t.Errorf("not found expected headers: %+v", r.Header)
		}

		if len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != "chunked" {
			t.Errorf("not found expected transfer encoding: %+v", r.TransferEncoding)
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	req := BufferRequest(specs.HttpMethodPost, specs.MustParseUrl(server.URL), specs.ContentTypePlain, requestBody)
	req.Header().Set("x-hello-world", "xyz-123")
	req.Header().Set("Transfer-Encoding", "chunked")

	resp, err := DefaultTransport().RoundTrip(
		context.Background(), req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

func TestTransport_PostChunkedTransferEncodingRequest(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	requestBody := []byte(`{"key": "value"}`)

	go func() {
		var conn net.Conn
		for {
			conn, err = listener.Accept()

			select {
			case <-ctx.Done():
			default:
			}

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				t.Fatal(err)
			}
			break
		}

		reader := bufio.NewReader(conn)
		req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), reader, 1024, 8024)
		if err != nil {
			t.Fatal(err)
		}

		if req.Header().Get("X-Hello-World") != "xyz-123" ||
			req.Header().Get("Transfer-Encoding") != "chunked" {
			t.Errorf("not found expected headers: %+v", req.Header())
		}

		cr := encoding.NewChunkedReader(reader)
		b, err := io.ReadAll(cr)
		if err != nil {
			t.Fatalf("Read all: %s", err)
		}
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}

		server_ops.WriteResponseHead(conn, true, specs.StatusCodeOK, specs.NewHeader(func(header *specs.Header) {
			header.Set("Content-Length", "8")
		}))
		conn.Write([]byte("received"))
		conn.Close()
	}()

	defer func() {
		listener.Close()
		cancel()
	}()

	url := specs.MustParseUrl("http://" + listener.Addr().String())
	req := BufferRequest(specs.HttpMethodPost, url, specs.ContentTypePlain, requestBody)
	req.Header().Set("Transfer-Encoding", "chunked")
	req.Header().Set("x-hello-world", "xyz-123")

	resp, err := DefaultTransport().RoundTrip(
		context.Background(), req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

// Test content encoding

func TestTransport_GzipEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := gzip.NewWriter(&cacheBuf)
		cw.Write(testContent)
		cw.Close()

		body := cacheBuf.Bytes()

		header := specs.NewHeader()
		header.Set("Content-Encoding", "gzip")
		header.Set("Content-Length", strconv.Itoa(len(body)))
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

func TestTransport_DeflateEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := zlib.NewWriter(&cacheBuf)
		cw.Write(testContent)
		cw.Close()

		body := cacheBuf.Bytes()

		header := specs.NewHeader()
		header.Set("Content-Encoding", "deflate")
		header.Set("Content-Length", strconv.Itoa(len(body)))
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "deflate" {
		t.Errorf("expected deflate encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

func TestTransport_BrotliEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := brotli.NewWriter(&cacheBuf)
		cw.Write(testContent)
		cw.Close()

		body := cacheBuf.Bytes()
		header := specs.NewHeader()
		header.Set("Content-Encoding", "br")
		header.Set("Content-Length", strconv.Itoa(len(body)))
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "br" {
		t.Errorf("expected br encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

// Test combined encoding and chunked

func TestTransport_ChunkedAndGzipEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := encoding.NewChunkedWriter(&cacheBuf)
		ew := gzip.NewWriter(cw)
		ew.Write(testContent)
		ew.Close()
		cw.Close()

		body := cacheBuf.Bytes()

		header := specs.NewHeader()
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Content-Encoding", "gzip")
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

func TestTransport_ChunkedAndDeflateEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := encoding.NewChunkedWriter(&cacheBuf)
		ew := zlib.NewWriter(cw)
		ew.Write(testContent)
		ew.Close()
		cw.Close()

		body := cacheBuf.Bytes()
		header := specs.NewHeader()
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Content-Encoding", "deflate")
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "deflate" {
		t.Errorf("expected deflate encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

func TestTransport_ChunkedAndBrotliEncoding(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		var cacheBuf bytes.Buffer
		cw := encoding.NewChunkedWriter(&cacheBuf)
		ew := brotli.NewWriter(cw)
		ew.Write(testContent)
		ew.Close()
		cw.Close()

		body := cacheBuf.Bytes()
		header := specs.NewHeader()
		header.Set("Transfer-Encoding", "chunked")
		header.Set("Content-Encoding", "br")
		return specs.StatusCodeOK, header, body
	})
	defer closeServer()

	resp, err := DefaultTransport().RoundTrip(context.Background(), specs.HttpMethodGet, *url, specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("Content-Encoding") != "br" {
		t.Errorf("expected br encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	checkResponseBody(t, resp, testContent)
}

// Test TLS

func TestTransport_GetRequestTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-hello-world", "xyz-123")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	url := "https://" + server.Listener.Addr().String()

	transport := DefaultTransport()
	transport.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *specs.MustParseUrl(url), specs.NewHeader(), nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Hello-World") != "xyz-123" ||
		resp.Header().Get("Content-Encoding") != "" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, []byte("OK"))
}

func TestTransport_PostRequestTLS(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Hello-World") != "xyz-123" ||
			r.Header.Get("Content-Length") != strconv.Itoa(len(requestBody)) {
			t.Error("not found expected headers", r.Header)
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	url := "https://" + server.Listener.Addr().String()
	transport := DefaultTransport()
	transport.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	req := BufferRequest(specs.HttpMethodPost, specs.MustParseUrl(url), specs.ContentTypePlain, requestBody)
	req.Header().Set("x-hello-world", "xyz-123")

	resp, err := transport.RoundTrip(
		context.Background(), req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

// Test proxy

func TestTransport_Sock5Proxy(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	proxyServer, err := socks5.New(&socks5.Config{})
	if err != nil {
		t.Fatal(err)
	}

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var connectedProxy atomic.Bool

	go func() {
		var conn net.Conn
		var err error
		for {
			conn, err = proxyListener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				t.Error(err)
			}
			break
		}
		connectedProxy.Store(true)
		if err := proxyServer.ServeConn(conn); err != nil {
			t.Error(err)
		}
	}()

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	transport := DefaultTransport()
	transport.Proxy = FixedProxyUrl(specs.MustParseUrl("socks5://" + proxyListener.Addr().String()))

	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

func TestTransport_Sock5ProxyAuth(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	proxyServer, err := socks5.New(&socks5.Config{
		Credentials: &socks5.StaticCredentials{
			"username": "password",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var connectedProxy atomic.Bool

	go func() {
		var conn net.Conn
		var err error
		for {
			conn, err = proxyListener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
				t.Error(err)
			}
			break
		}
		connectedProxy.Store(true)
		if err := proxyServer.ServeConn(conn); err != nil {
			t.Error(err)
		}
	}()

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	transport := DefaultTransport()
	transport.Proxy = FixedProxyUrl(specs.MustParseUrl("socks5h://username:password@" + proxyListener.Addr().String()))

	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

func TestTransport_HttpProxy(t *testing.T) {
	var connectedProxy atomic.Bool
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, proxyUrl := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" ||
			req.Header().Get("Host") != "test.org:80" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}
		connectedProxy.Store(true)

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	transport := DefaultTransport()
	transport.Proxy = FixedProxyUrl(proxyUrl)

	url := specs.MustParseUrl("http://test.org/")
	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

func TestTransport_HttpProxyAuth(t *testing.T) {
	var connectedProxy atomic.Bool
	testContent := []byte("Content\nEncoding 1234567890")
	closeServer, proxyUrl := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" ||
			req.Header().Get("Proxy-Authorization") != specs.BasicAuthHeader("username", "password") ||
			req.Header().Get("Host") != "test.org:80" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}
		connectedProxy.Store(true)

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	proxyUrl.Username = "username"
	proxyUrl.Password = "password"

	transport := DefaultTransport()
	transport.Proxy = FixedProxyUrl(proxyUrl)

	url := specs.MustParseUrl("http://test.org/")
	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

func TestTransport_HttpsProxy(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")

	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	var connectedProxy atomic.Bool

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		destConn, err := net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		fmt.Fprint(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")

		transfer := func(dst net.Conn, src net.Conn) {
			defer dst.Close()
			defer src.Close()
			io.Copy(dst, src)
		}

		go transfer(destConn, clientConn)
		go transfer(clientConn, destConn)

		connectedProxy.Store(true)
	}))
	defer proxyServer.Close()

	proxyUrl := specs.MustParseUrl(proxyServer.URL)

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	transport := DefaultTransport()
	proxyUrl.Scheme = "https"
	transport.Proxy = FixedProxyUrl(proxyUrl)

	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

func TestTransport_HttpsProxyAuth(t *testing.T) {
	testContent := []byte("Content\nEncoding 1234567890")

	closeServer, url := newTestServer(func(req Request) (specs.StatusCode, *specs.Header, []byte) {
		if req.Header().Get("X-ping") != "xyz-123" {
			t.Errorf("not found expected headers, %+v", req.Header())
		}

		header := specs.NewHeader()
		header.Set("Content-Length", strconv.Itoa(len(testContent)))
		header.Set("x-pong", "xyz-321")

		return specs.StatusCodeOK, header, testContent
	})
	defer closeServer()

	var connectedProxy atomic.Bool

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Proxy-Authorization") != specs.BasicAuthHeader("usern", "pass") {
			user, pass, _ := specs.ParseBasicAuthHeader(r.Header.Get("Proxy-Authorization"))
			t.Errorf("invalid creds: %s : %s", user, pass)
			http.Error(w, "Invalid creds", http.StatusProxyAuthRequired)
			return
		}

		destConn, err := net.Dial("tcp", r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		fmt.Fprint(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")

		transfer := func(dst net.Conn, src net.Conn) {
			defer dst.Close()
			defer src.Close()
			io.Copy(dst, src)
		}

		go transfer(destConn, clientConn)
		go transfer(clientConn, destConn)

		connectedProxy.Store(true)
	}))
	defer proxyServer.Close()

	proxyUrl := specs.MustParseUrl(proxyServer.URL)

	header := specs.NewHeader()
	header.Set("x-ping", "xyz-123")

	transport := DefaultTransport()
	proxyUrl.Scheme = "https"
	proxyUrl.Username = "usern"
	proxyUrl.Password = "pass"
	transport.Proxy = FixedProxyUrl(proxyUrl)

	resp, err := transport.RoundTrip(context.Background(), specs.HttpMethodGet, *url, header, nil)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header().Get("X-Pong") != "xyz-321" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	checkResponseBody(t, resp, testContent)

	if !connectedProxy.Load() {
		t.Fatal("not was connected to proxy server")
	}
}

// Test Expectation

func TestTransport_Expect100ContinueRaw(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	ctx, cancel := context.WithCancel(context.Background())
	listener, err := serveTcpTest(ctx, func(conn net.Conn) {
		bufioReader := bufio.NewReader(conn)
		req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), bufioReader, 1024, 8*1024)
		if err != nil {
			t.Error(err)
		}

		if req.Header().Get("Expect") != "100-continue" ||
			req.Header().Get("Content-Length") != "16" {
			t.Errorf("unexpected headers, %+v", req.Header())
		}

		conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		_, err = bufioReader.Peek(2)
		if err != nil {
			var neterr net.Error
			if !(errors.As(err, &neterr) && neterr.Timeout()) {
				t.Error(err)
			}
		}
		conn.SetReadDeadline(time.Time{})

		_, err = conn.Write(responseContinueBuf)
		if err != nil {
			t.Error(err)
		}

		buf := make([]byte, len(requestBody))
		_, err = io.ReadFull(bufioReader, buf)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(buf, requestBody) {
			t.Errorf("expect '%s', got '%s'", requestBody, buf)
		}

		header := specs.NewHeader()
		header.Set("Content-Length", "4")
		_, err = server_ops.WriteResponseHead(conn, true, specs.StatusCodeOK, header)
		if err != nil {
			t.Error(err)
		}

		_, err = conn.Write([]byte("okay"))
		if err != nil {
			t.Error(err)
		}

	})
	defer func() {
		cancel()
		listener.Close()
	}()

	url := specs.MustParseUrl("http://" + listener.Addr().String())
	req := BufferRequest(specs.HttpMethodPost, url, specs.ContentTypePlain, requestBody)
	req.Header().Set("Expect", "100-continue")

	resp, err := DefaultTransport().RoundTrip(ctx, req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("okay"))
}

func TestTransport_Expect100ContinueHttp(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Expect") != "100-continue" ||
			r.Header.Get("Content-Length") != strconv.Itoa(len(requestBody)) ||
			r.Header.Get("Content-Type") != specs.ContentTypePlain {
			t.Errorf("not found expected headers: %+v", r.Header)
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("okay"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := BufferRequest(specs.HttpMethodPost, url, specs.ContentTypePlain, requestBody)
	req.Header().Set("Expect", "100-continue")

	resp, err := DefaultTransport().RoundTrip(context.Background(), req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("okay"))
}

// Test Hijack

func TestTransport_Hijack(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	ctx, cancel := context.WithCancel(context.Background())
	listener, err := serveTcpTest(ctx, func(conn net.Conn) {
		// Reading
		bufioReader := bufio.NewReader(conn)
		req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), bufioReader, 1024, 8*1024)
		if err != nil {
			t.Error(err)
		}

		if req.Header().Get("Content-Length") != "16" {
			t.Errorf("unexpected headers, %+v", req.Header())
		}

		buf := make([]byte, len(requestBody))
		_, err = io.ReadFull(bufioReader, buf)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(buf, requestBody) {
			t.Errorf("expect '%s', got '%s'", requestBody, buf)
		}

		// Writing
		header := specs.NewHeader()
		header.Set("Content-Length", "4")
		_, err = server_ops.WriteResponseHead(conn, true, specs.StatusCodeOK, header)
		if err != nil {
			t.Error(err)
		}

		_, err = conn.Write([]byte("okay"))
		if err != nil {
			t.Error(err)
		}

		// After hijack

		buf = make([]byte, 4)
		_, err = io.ReadFull(bufioReader, buf)
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(buf, []byte("ping")) {
			t.Errorf("unexpected pong response '%s'", buf)
		}

		_, err = conn.Write([]byte("pong"))
		if err != nil {
			t.Error(err)
		}
	})
	defer func() {
		cancel()
		listener.Close()
	}()

	url := specs.MustParseUrl("http://" + listener.Addr().String())
	req := BufferRequest(specs.HttpMethodPost, url, specs.ContentTypePlain, requestBody)
	hijacker, ctx := WithTransportHijacker(ctx)

	resp, err := DefaultTransport().RoundTrip(ctx, req.Method(), req.Url(), req.Header(), req.(BodyWriter))

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Error("invalid status code:", resp.StatusCode())
	}

	body := resp.Body()
	if body == nil {
		t.Error("response body is nil")
	}
	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Error("read all:", err)
	}

	if !bytes.Equal(data, []byte("okay")) {
		t.Error("invalid response:", string(data))
	}

	conn := hijacker.Conn
	if conn == nil {
		t.Error("conn not hijacked")
	}

	_, err = conn.Write([]byte("ping"))
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, 4)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(buf, []byte("pong")) {
		t.Errorf("unexpected pong response '%s'", buf)
	}
}
