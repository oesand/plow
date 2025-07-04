package giglet

import (
	"bytes"
	"context"
	"fmt"
	"github.com/oesand/giglet/specs"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_GetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	resp, err := DefaultClient().Make(NewRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("OK"))
}

func TestClient_PostRequest(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	req := NewBufferRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL), requestBody, specs.ContentTypePlain)
	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

func TestClient_Redirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/final", http.StatusFound)
		} else if r.URL.Path == "/final" {
			fmt.Fprint(w, "Final Destination")
		} else {
			fmt.Fprint(w, "Invalid flow")
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	resp, err := DefaultClient().Make(NewRequest(specs.HttpMethodGet, url))
	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("Final Destination"))
}

func TestClient_TooManyRedirects(t *testing.T) {
	maxRedirectCount := 5
	var serverVisits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
		serverVisits++
	}))
	defer server.Close()

	client := &Client{
		MaxRedirectCount: maxRedirectCount,
	}

	_, err := client.Make(NewRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err == nil || err.Error() != "giglet/redirect: too many redirects" {
		t.Errorf("invalid error: %s, expected 'too many redirects'", err)
	}

	if serverVisits != maxRedirectCount+1 {
		t.Errorf("invalid server count visits: %d, expected %d", serverVisits, maxRedirectCount)
	}
}

func TestClient_RedirectMissingLocationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	_, err := DefaultClient().Make(NewRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err == nil || err.Error() != "giglet/redirect: empty Location header" {
		t.Errorf("expected error on empty Location header, got %v", err)
	}
}

func TestClient_RedirectInvalidLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", ":bad_url")
		w.Header().Set("Lol", ":bad_url")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	_, err := DefaultClient().Make(NewRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err == nil || !strings.Contains(err.Error(), "cannot parse location") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestClient_ChunkedTransferEncoding(t *testing.T) {
	closeServer := newTestServer(func(ctx context.Context, req Request) Response {
		return NewTextResponse("Answered chunked", specs.ContentTypePlain)
	})
	defer closeServer()

	req := NewRequest(specs.HttpMethodGet, specs.MustParseUrl("http://127.0.0.1:80"))
	//req.Header().Set("Transfer-Encoding", "chunked")

	client := DefaultClient()
	resp, err := client.Make(req)

	if err != nil {
		t.Fatal("req:", err)
	}

	t.Logf("<%d>: %v \n", resp.StatusCode(), resp.Header())

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	data, err := io.ReadAll(body)
	t.Logf("data: %s \n", data)

	http.Serve()

	// TODO : fix invalid Content-Length with encoding

	//http.Serve()

	/*
		closeServer := newTestServer(func(ctx context.Context, req Request) Response {
			return NewTextResponse("Answered chunked", specs.ContentTypePlain)
		})
		defer closeServer()

		req := NewRequest(specs.HttpMethodGet, specs.MustParseUrl("http://127.0.0.1:80"))
		//req.Header().Set("Transfer-Encoding", "chunked")

		client := DefaultClient()
		resp, err := client.Make(req)
		if err != nil {
			t.Fatal("req:", err)
		}

		body := resp.Body()
		if body == nil {
			t.Fatal("response body is nil")
		}

		defer body.Close()

		fmt.Printf("Transfer-Encoding: %s \n", resp.Header().Get("Transfer-Encoding"))
		fmt.Printf("Content-Encoding: %s \n", resp.Header().Get("Content-Encoding"))
		fmt.Printf("Data: %v \n", resp.Header())

		data, err := io.ReadAll(body)
		if err != nil {
			t.Fatal("read all:", err)
		}

		if resp.StatusCode() != specs.StatusCodeOK {
			t.Fatal("invalid status code:", resp.StatusCode(), ", body:", string(data))
		}

		if !bytes.Equal(data, []byte("Answered chunked")) {
			t.Error("invalid response:", string(data))
		} else {
			t.Logf("valid data: %s \n", data)
		}

	*/
}
