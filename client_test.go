package giglet

import (
	"bytes"
	"fmt"
	"github.com/oesand/giglet/specs"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestClient_GetRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-hello-world", "xyz-123")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	resp, err := DefaultClient().Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
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

func TestClient_PostRequest(t *testing.T) {
	requestBody := []byte(`{"key": "value"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Hello-World") != "xyz-123" ||
			r.Header.Get("Content-Length") != strconv.Itoa(len(requestBody)) ||
			r.Header.Get("x-Type") != "json" {
			t.Errorf("not found expected headers: %+v", r.Header)
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, requestBody) {
			t.Errorf("expected %s, got %s", string(requestBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	req := BufferRequest(specs.HttpMethodPost, specs.MustParseUrl(server.URL), requestBody, specs.ContentTypePlain)
	req.Header().Set("x-type", "json")
	req.Header().Set("x-hello-world", "xyz-123")

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
	resp, err := DefaultClient().Make(EmptyRequest(specs.HttpMethodGet, url))
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

	_, err := client.Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
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

	_, err := DefaultClient().Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
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

	_, err := DefaultClient().Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err == nil || !strings.Contains(err.Error(), "cannot parse location") {
		t.Errorf("expected parse error, got %v", err)
	}
}

// Test all Requests

func TestClient_PostAnyRequest(t *testing.T) {
	tests := []struct {
		name     string
		request  func(url *specs.Url) ClientRequest
		wantBody []byte
	}{
		{
			name: "TextRequest",
			request: func(url *specs.Url) ClientRequest {
				return TextRequest(specs.HttpMethodPost, url, "text-request-body", specs.ContentTypePlain)
			},
			wantBody: []byte("text-request-body"),
		},
		{
			name: "BufferRequest",
			request: func(url *specs.Url) ClientRequest {
				return BufferRequest(specs.HttpMethodPost, url, []byte("buffer-request-body"), specs.ContentTypeRaw)
			},
			wantBody: []byte("buffer-request-body"),
		},
		{
			name: "StreamRequest",
			request: func(url *specs.Url) ClientRequest {
				var buf bytes.Buffer
				buf.WriteString("stream-request-body")
				return StreamRequest(specs.HttpMethodPost, url, &buf, specs.ContentTypeRaw, int64(buf.Len()))
			},
			wantBody: []byte("stream-request-body"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				if !bytes.Equal(b, tt.wantBody) {
					t.Errorf("expected %s, got %s", string(tt.wantBody), string(b))
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			req := tt.request(specs.MustParseUrl(server.URL))
			resp, err := DefaultClient().Make(req)
			if err != nil {
				t.Fatal("req:", err)
			}

			if resp.StatusCode() != specs.StatusCodeOK {
				t.Fatal("invalid status code:", resp.StatusCode())
			}
		})
	}
}

// Test Client.Jar

func TestClient_ClientWithJar(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Cookies())
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Jar = specs.NewCookieJar()
	client.Jar.SetCookie("127.0.0.1", specs.Cookie{
		Name:  cookieName,
		Value: cookieValue,
	})

	resp, err := client.Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

func TestClient_ClientWithJarAndAlreadyHasCookie(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Jar = specs.NewCookieJar()
	client.Jar.SetCookie("127.0.0.1", specs.Cookie{
		Name:  cookieName,
		Value: "not-valid-value",
	})

	req := EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL))
	req.Header().SetCookieValue(cookieName, cookieValue)

	resp, err := client.Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

// Test Client.Header headers

func TestClient_ClientWithHeader(t *testing.T) {
	headerName, headerValue := "X-Header-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerName) != headerValue {
			t.Errorf("not found expected headers, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Header = specs.NewHeader()
	client.Header.Set(headerName, headerValue)

	resp, err := client.Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

func TestClient_ClientWithHeaderAndAlreadyHasHeader(t *testing.T) {
	headerName, headerValue := "X-Header-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerName) != headerValue {
			t.Errorf("not found expected headers, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Header = specs.NewHeader()
	client.Header.Set(headerName, "not-valid-value")

	req := EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL))
	req.Header().Set(headerName, headerValue)

	resp, err := client.Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

// Test Client.Header cookies

func TestClient_ClientWithHeaderCookies(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Cookies())
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Header = specs.NewHeader()
	client.Header.SetCookieValue(cookieName, cookieValue)

	resp, err := client.Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

func TestClient_ClientWithHeaderCookiesAndAlreadyHasCookie(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()
	client.Header = specs.NewHeader()
	client.Header.SetCookieValue(cookieName, "not-valid-value")

	req := EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL))
	req.Header().SetCookieValue(cookieName, cookieValue)

	resp, err := client.Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

// Test combined Client.Jar Client.Header.Cookies

func TestClient_ClientWithHeaderCookiesAndJar(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()

	client.Jar = specs.NewCookieJar()
	client.Jar.SetCookie("127.0.0.1", specs.Cookie{
		Name:  cookieName,
		Value: cookieValue,
	})

	client.Header = specs.NewHeader()
	client.Header.SetCookieValue(cookieName, "not-valid-value")

	resp, err := client.Make(EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL)))
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}

func TestClient_ClientWithHeaderCookiesAndJarAndAlreadyHasCookie(t *testing.T) {
	cookieName, cookieValue := "X-Cookie-Name", "xyz-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie(cookieName)
		if cookie.Name != cookieName || cookie.Value != cookieValue {
			t.Errorf("not found expected cookies, %+v", r.Header)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := DefaultClient()

	client.Jar = specs.NewCookieJar()
	client.Jar.SetCookie("127.0.0.1", specs.Cookie{
		Name:  cookieName,
		Value: "not-valid-value",
	})

	client.Header = specs.NewHeader()
	client.Header.SetCookieValue(cookieName, "not-valid-value")

	req := EmptyRequest(specs.HttpMethodGet, specs.MustParseUrl(server.URL))
	req.Header().SetCookieValue(cookieName, cookieValue)

	resp, err := client.Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}
}
