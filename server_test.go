package giglet

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"github.com/andybalholm/brotli"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/encoding"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/mock"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestServer_GetRequest(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay", func(resp Response) {
			resp.Header().Set("x-hello-world", "xyz-123")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := "http://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("X-Hello-World") != "xyz-123" ||
		resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

func TestServer_PostRequest(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Content-Length") != strconv.Itoa(len(requestBody)) {
			t.Error("not found expected headers")
		}

		b, _ := io.ReadAll(request.Body())
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay", func(resp Response) {
			resp.Header().Set("x-hello-world", "321-xyz")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := "http://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("X-Hello-World", "xyz-123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("X-Hello-World") != "321-xyz" ||
		resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

func TestServer_SendAnyResponse(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantBody []byte
	}{
		{
			name:     "TextResponse",
			response: TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "text response"),
			wantBody: []byte("text response"),
		},
		{
			name:     "BufferResponse",
			response: BufferResponse(specs.StatusCodeOK, specs.ContentTypePlain, []byte("buffer response")),
			wantBody: []byte("buffer response"),
		},
		{
			name:     "StreamResponse",
			response: StreamResponse(specs.StatusCodeOK, specs.ContentTypePlain, bytes.NewReader([]byte("stream response")), 15),
			wantBody: []byte("stream response"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
				return tt.response
			}))

			listener, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			go server.Serve(listener)

			url := "http://" + listener.Addr().String()

			client := &http.Client{Transport: &http.Transport{}}
			req, _ := http.NewRequest("GET", url, nil)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal("req:", err)
			}

			checkHttpResponseBody(t, resp, tt.wantBody)
		})
	}
}

// Test chunked encoding

func TestServer_ChunkedTransferEncodingTwoWays(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		req := request.(*server.HttpRequest)
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Transfer-Encoding") != "chunked" {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		if !req.Chunked {
			t.Error("chunked flag not set int request")
		}

		body := req.Body()
		if body == nil {
			t.Fatal("request body is nil")
		}

		data, err := io.ReadAll(body)
		if err != nil {
			t.Fatal("read all:", err)
		}

		if !bytes.Equal(data, []byte("request encoded")) {
			t.Error("invalid request:", string(data))
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("x-hello-world", "xyz-123")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("x-hello-world", "xyz-123")

	encodedContent := bytes.Buffer{}
	cw := encoding.NewChunkedWriter(&encodedContent)
	_, err = cw.Write([]byte("request encoded"))
	if err == nil {
		err = cw.Close()
	}
	if err != nil {
		t.Fatal("fail to write chunked:", err)
	}

	resp, _, err := newTestClientSend(specs.HttpMethodPost, url, header, encodedContent.Bytes())
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("X-Hello-World") != "xyz-123" ||
		resp.Header().Get("Transfer-Encoding") != "chunked" ||
		resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := httputil.NewChunkedReader(body)
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("response encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_ChunkedTransferEncodingResponse(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("x-hello-world", "xyz-123")
			resp.Header().Set("Transfer-Encoding", "chunked")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("X-Hello-World") != "xyz-123" ||
		resp.Header().Get("Transfer-Encoding") != "chunked" ||
		resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := httputil.NewChunkedReader(body)
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("response encoded")) {
		t.Error("invalid response:", string(data))
	}
}

// Test content encoding

func TestServer_GzipEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingGzip {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay encoded")
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingGzip)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingGzip {
		t.Errorf("expected gzip encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	if resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	contentLength, err := strconv.Atoi(resp.Header().Get("Content-Length"))
	if err != nil {
		t.Fatalf("invalid content length header: %s", resp.Header().Get("Content-Length"))
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := io.LimitReader(body, int64(contentLength))
	reader, err = gzip.NewReader(reader)
	if err != nil {
		t.Fatalf("encoder err: %s", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("okay encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_DeflateEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingDeflate {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay encoded")
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingDeflate)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingDeflate {
		t.Errorf("expected deflate encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	contentLength, err := strconv.Atoi(resp.Header().Get("Content-Length"))
	if err != nil {
		t.Fatalf("invalid content length header: %s", resp.Header().Get("Content-Length"))
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := io.LimitReader(body, int64(contentLength))
	reader = flate.NewReader(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("okay encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_BrotliEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingBrotli {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay encoded")
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingBrotli)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingBrotli {
		t.Errorf("expected brotli encoding, got %s", resp.Header().Get("Content-Encoding"))
	}

	contentLength, err := strconv.Atoi(resp.Header().Get("Content-Length"))
	if err != nil {
		t.Fatalf("invalid content length header: %s", resp.Header().Get("Content-Length"))
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := io.LimitReader(body, int64(contentLength))
	reader = brotli.NewReader(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("okay encoded")) {
		t.Error("invalid response:", string(data))
	}
}

// Test combined encoding and chunked transfer

func TestServer_GzipEncodingAndChunkedTransferEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingGzip {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("Transfer-Encoding", "chunked")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingGzip)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingGzip ||
		resp.Header().Get("Transfer-Encoding") != "chunked" ||
		resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := httputil.NewChunkedReader(body)
	reader, err = gzip.NewReader(reader)
	if err != nil {
		t.Fatalf("encoder err: %s", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("response encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_DeflateEncodingAndChunkedTransferEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingDeflate {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("Transfer-Encoding", "chunked")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingDeflate)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingDeflate ||
		resp.Header().Get("Transfer-Encoding") != "chunked" ||
		resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := httputil.NewChunkedReader(body)
	reader = flate.NewReader(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("response encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_BrotliEncodingAndChunkedTransferEncoding(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Accept-Encoding") != specs.ContentEncodingBrotli {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("Transfer-Encoding", "chunked")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Accept-Encoding", specs.ContentEncodingBrotli)
	header.Set("x-hello-world", "xyz-123")

	resp, _, err := newTestClientSend(specs.HttpMethodGet, url, header, nil)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("Content-Encoding") != specs.ContentEncodingBrotli ||
		resp.Header().Get("Transfer-Encoding") != "chunked" ||
		resp.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := httputil.NewChunkedReader(body)
	reader = brotli.NewReader(reader)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("response encoded")) {
		t.Error("invalid response:", string(data))
	}
}

func TestServer_GzipEncodingAndChunkedTransferEncoding_ByHttpTestClient(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("Accept-Encoding") != specs.ContentEncodingGzip {
			t.Errorf("not found expected headers, %+v", request.Header())
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "response encoded", func(resp Response) {
			resp.Header().Set("Transfer-Encoding", "chunked")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := "http://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if len(resp.TransferEncoding) != 1 || resp.TransferEncoding[0] != "chunked" {
		t.Errorf("invalid transfer encoding %+v", resp.TransferEncoding)
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("response encoded"))
}

// Test other functionality

func TestServer_Hijack(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)
	responseBody := []byte(`response okay`)

	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Content-Length") != strconv.Itoa(len(requestBody)) {
			t.Error("not found expected headers")
		}

		b, _ := io.ReadAll(request.Body())
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}

		request.Hijack(func(ctx context.Context, conn net.Conn) {
			var rb = make([]byte, 4)
			_, err := conn.Read(rb)
			if err != nil {
				t.Error("fail to read after hijack, err:", err)
			}
			if !bytes.Equal(rb, []byte("ping")) {
				t.Error("not found expected hijack ping, actual:", rb)
			}

			conn.Write([]byte("pong"))
		})

		return BufferResponse(specs.StatusCodeOK, specs.ContentTypeRaw, responseBody, func(resp Response) {
			resp.Header().Set("x-hello-world", "321-xyz")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	header := specs.NewHeader()
	header.Set("Content-Length", strconv.Itoa(len(requestBody)))
	header.Set("x-hello-world", "xyz-123")

	resp, conn, err := newTestClientSend(specs.HttpMethodPost, url, header, requestBody)
	if err != nil {
		t.Fatal("req:", err)
	}

	if code := resp.StatusCode(); code != specs.StatusCodeOK {
		t.Fatal("invalid status code:", code)
	}

	if resp.Header().Get("X-Hello-World") != "321-xyz" ||
		resp.Header().Get("Content-Length") != strconv.Itoa(len(responseBody)) ||
		resp.Header().Get("Content-Type") != specs.ContentTypeRaw {
		t.Errorf("not found expected headers, %+v", resp.Header())
	}

	contentLength, err := strconv.Atoi(resp.Header().Get("Content-Length"))
	if err != nil {
		t.Fatalf("invalid content length header: %s", resp.Header().Get("Content-Length"))
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	reader := io.LimitReader(body, int64(contentLength))
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, responseBody) {
		t.Error("invalid response:", string(data))
	}

	conn.Write([]byte("ping"))

	var rb = make([]byte, 4)
	_, err = conn.Read(rb)
	if err != nil {
		t.Error("fail to read from hijacked server, err:", err)
	}
	if !bytes.Equal(rb, []byte("pong")) {
		t.Error("not found expected hijack pong, actual:", rb)
	}
}

func TestServer_FilterConn(t *testing.T) {
	var wasChecked atomic.Bool
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay")
	}))
	server.FilterConn = func(addr net.Addr) bool {
		wasChecked.Store(true)
		return false
	}

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := specs.MustParseUrl(listener.Addr().String())

	address := client.HostPort(url.Host, url.Port)

	conn, err := defaultDialer.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial err: %s", err)
	}

	_, err = conn.Read(make([]byte, 1))
	if err != io.EOF {
		t.Error("conn not closed, err:", err)
	}

	if !wasChecked.Load() {
		t.Fatal("FilterConn not was triggered")
	}
}

// Test TLS

func TestServer_GetRequestTLS(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay", func(resp Response) {
			resp.Header().Set("x-hello-world", "xyz-123")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.ServeTLSRaw(listener, mock.NewTlsCert())

	url := "https://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("X-Hello-World") != "xyz-123" ||
		resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

func TestServer_PostRequestTLS(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Content-Length") != strconv.Itoa(len(requestBody)) {
			t.Error("not found expected headers")
		}

		b, _ := io.ReadAll(request.Body())
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay", func(resp Response) {
			resp.Header().Set("x-hello-world", "321-xyz")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.ServeTLSRaw(listener, mock.NewTlsCert())

	url := "https://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("X-Hello-World", "xyz-123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("X-Hello-World") != "321-xyz" ||
		resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

func TestServer_PanicHandling(t *testing.T) {
	var panicHandled atomic.Bool
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		panic("test panic")
	}))
	server.ErrorHandler = ErrorHandlerFunc(func(ctx context.Context, conn net.Conn, err any) {
		ShortResponseWriter(specs.StatusCodeInternalServerError, "panic handled").WriteTo(conn)
		panicHandled.Store(true)
	})

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := "http://" + listener.Addr().String()
	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}
	if resp.StatusCode != int(specs.StatusCodeInternalServerError) {
		t.Errorf("expected status code 500, got %d", resp.StatusCode)
	}

	body := resp.Body
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, []byte("panic handled")) {
		t.Error("invalid response:", string(data))
	}

	if !panicHandled.Load() {
		t.Fatal("panic not handled")
	}
}

func TestServer_RequestBodyTooLarge(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "ok")
	}))
	server.MaxBodySize = 4

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)

	url := "http://" + listener.Addr().String()
	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte("this is a test body that is too large")))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}
	if resp.StatusCode != int(specs.StatusCodeRequestEntityTooLarge) {
		t.Errorf("expected status code 413, go %d", resp.StatusCode)
	}
}
