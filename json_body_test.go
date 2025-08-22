package plow

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/oesand/plow/specs"
)

// Test data structures for JSON testing
type testUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type testProduct struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func TestJsonRequest(t *testing.T) {
	expectedUser := testUser{ID: 1, Name: "John Doe", Age: 30}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedBody, _ := json.Marshal(expectedUser)

		if r.Header.Get("Content-Type") != specs.ContentTypeJson {
			t.Errorf("expected Content-Type %s, got %s", specs.ContentTypeJson, r.Header.Get("Content-Type"))
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, expectedBody) {
			t.Errorf("expected %s, got %s", string(expectedBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req, err := JsonRequest(specs.HttpMethodPost, url, expectedUser)
	if err != nil {
		t.Fatal("JsonRequest failed:", err)
	}

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

func TestJsonRequestWithComplexData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedProduct := testProduct{SKU: "ABC123", Price: 29.99, Stock: 100}
		expectedBody, _ := json.Marshal(expectedProduct)

		if r.Header.Get("Content-Type") != specs.ContentTypeJson {
			t.Errorf("expected Content-Type %s, got %s", specs.ContentTypeJson, r.Header.Get("Content-Type"))
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, expectedBody) {
			t.Errorf("expected %s, got %s", string(expectedBody), string(b))
		}
		w.Write([]byte("product received"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	product := testProduct{SKU: "ABC123", Price: 29.99, Stock: 100}

	req, err := JsonRequest(specs.HttpMethodPut, url, product)
	if err != nil {
		t.Fatal("JsonRequest failed:", err)
	}

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	checkResponseBody(t, resp, []byte("product received"))
}

func TestReadJsonFromRequest(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		user, err := ReadJson[testUser](request)
		if err != nil {
			t.Error("ReadJson failed:", err)
			return TextResponse(specs.StatusCodeBadRequest, specs.ContentTypePlain, "error")
		}

		expectedUser := testUser{ID: 2, Name: "Jane Smith", Age: 25}
		if !reflect.DeepEqual(*user, expectedUser) {
			t.Errorf("read invalid user = %+v, want %+v", *user, expectedUser)
		}

		return TextResponse(specs.StatusCodeOK, specs.ContentTypePlain, "okay")
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go server.Serve(listener)

	url := "http://" + listener.Addr().String()
	user := testUser{ID: 2, Name: "Jane Smith", Age: 25}
	userJSON, _ := json.Marshal(user)

	resp, err := http.Post(url, specs.ContentTypeJson, bytes.NewReader(userJSON))
	if err != nil {
		t.Fatal("request failed:", err)
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("unexpected Content-Type header: %s", resp.Header.Get("Content-Type"))
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}

func TestReadJsonFromClientResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := testUser{ID: 3, Name: "Bob Wilson", Age: 35}
		userJSON, _ := json.Marshal(user)

		w.Header().Set("Content-Type", specs.ContentTypeJson)
		w.Write(userJSON)
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := EmptyRequest(specs.HttpMethodGet, url)

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	user, err := ReadJson[testUser](resp)
	if err != nil {
		t.Fatal("ReadJson failed:", err)
	}

	expectedUser := testUser{ID: 3, Name: "Bob Wilson", Age: 35}
	if !reflect.DeepEqual(*user, expectedUser) {
		t.Errorf("read invalid user = %+v, want %+v", *user, expectedUser)
	}
}

func TestReadJsonWithInvalidContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", specs.ContentTypePlain)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := EmptyRequest(specs.HttpMethodGet, url)

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	_, err = ReadJson[testUser](resp)
	if err == nil {
		t.Error("expected error for invalid content type, got nil")
	}
	if err.Error() != "request Content-Type isn't "+specs.ContentTypeJson {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestReadJsonWithNilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", specs.ContentTypeJson)
		// No body written
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := EmptyRequest(specs.HttpMethodGet, url)

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	_, err = ReadJson[testUser](resp)
	if err == nil {
		t.Error("expected error for nil body, got nil")
	}
	if err.Error() != "missing body" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestReadJsonWithInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", specs.ContentTypeJson)
		w.Write([]byte(`{"id": "not a number", "name": "test"}`)) // Invalid JSON for TestUser
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := EmptyRequest(specs.HttpMethodGet, url)

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	_, err = ReadJson[testUser](resp)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestJsonResponse(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		user := testUser{ID: 4, Name: "Alice Brown", Age: 28}
		return JsonResponse(specs.StatusCodeOK, user)
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go server.Serve(listener)

	url := "http://" + listener.Addr().String()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	if resp.Header.Get("Content-Type") != specs.ContentTypeJson {
		t.Errorf("expected Content-Type %s, got %s", specs.ContentTypeJson, resp.Header.Get("Content-Type"))
	}

	expectedUser := testUser{ID: 4, Name: "Alice Brown", Age: 28}
	expectedJSON, _ := json.Marshal(expectedUser)
	checkHttpResponseBody(t, resp, expectedJSON)
}

func TestJsonResponseWithConfigure(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		product := testProduct{SKU: "XYZ789", Price: 99.99, Stock: 50}
		return JsonResponse(specs.StatusCodeOK, product, func(r Response) {
			r.Header().Set("X-Custom-Header", "test-value")
		})
	}))

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go server.Serve(listener)

	url := "http://" + listener.Addr().String()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal("request failed:", err)
	}

	if resp.Header.Get("Content-Type") != specs.ContentTypeJson {
		t.Errorf("expected Content-Type %s, got %s", specs.ContentTypeJson, resp.Header.Get("Content-Type"))
	}

	if resp.Header.Get("X-Custom-Header") != "test-value" {
		t.Errorf("expected X-Custom-Header 'test-value', got %s", resp.Header.Get("X-Custom-Header"))
	}

	expectedProduct := testProduct{SKU: "XYZ789", Price: 99.99, Stock: 50}
	expectedJSON, _ := json.Marshal(expectedProduct)
	checkHttpResponseBody(t, resp, expectedJSON)
}

func TestJsonResponseInstance(t *testing.T) {
	user := testUser{ID: 5, Name: "Charlie Davis", Age: 32}
	resp := JsonResponse(specs.StatusCodeOK, user)

	instance := resp.Instance()
	if instance == nil {
		t.Error("Instance() returned nil")
	}

	// Check if the instance is the same as the original user
	if !reflect.DeepEqual(instance, user) {
		t.Errorf("Instance() returned %+v, expected %+v", instance, user)
	}
}

func TestJsonRequestWithInvalidData(t *testing.T) {
	url := specs.MustParseUrl("http://example.com")

	// Create a channel which cannot be marshaled to JSON
	invalidData := make(chan int)

	_, err := JsonRequest(specs.HttpMethodPost, url, invalidData)
	if err == nil {
		t.Error("expected error for invalid JSON data, got nil")
	}
}
