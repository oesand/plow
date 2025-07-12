package giglet

import (
	"bytes"
	"context"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
)

// TODO : add tests

func TestServer_GetRequest(t *testing.T) {
	server := DefaultServer(func(ctx context.Context, request Request) Response {
		return NewTextResponse("okay", specs.ContentTypePlain, specs.StatusCodeOK, func(resp Response) {
			resp.Header().Set("x-hello-world", "xyz-123")
		})
	})

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer func() {
		go server.Shutdown()
	}()

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

	server := DefaultServer(func(ctx context.Context, request Request) Response {
		if request.Header().Get("X-Hello-World") != "xyz-123" ||
			request.Header().Get("Content-Length") != strconv.Itoa(len(requestBody)) {
			t.Error("not found expected headers")
		}

		b, _ := io.ReadAll(request.Body())
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}

		return NewTextResponse("okay", specs.ContentTypePlain, specs.StatusCodeOK)
	})

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer func() {
		go server.Shutdown()
	}()

	url := "http://" + listener.Addr().String()

	client := &http.Client{Transport: &http.Transport{}}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("X-Hello-World", "xyz-123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

// TODO : Cover Content-Encoding + Transfer-Encoding

// TODO : Cover hijack
// TODO : Cover FilterConn
// TODO : Cover TLS
