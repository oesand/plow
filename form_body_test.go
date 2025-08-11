package giglet

import (
	"bytes"
	"context"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"reflect"
	"strconv"
	"testing"
)

func TestPostFormRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedBody := []byte(`hello=world&key=value`)

		if r.Header.Get("Content-Length") != strconv.Itoa(len(expectedBody)) ||
			r.Header.Get("Content-Type") != specs.ContentTypeForm {
			t.Errorf("not found expected headers: %+v", r.Header)
		}

		b, _ := io.ReadAll(r.Body)
		if !bytes.Equal(b, expectedBody) {
			t.Errorf("expected %s, got %s", string(expectedBody), string(b))
		}
		w.Write([]byte("received"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := FormRequest(specs.HttpMethodPut, url, specs.Query{
		"hello": "world",
		"key":   "value",
	})

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

func TestReadForm(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		form, err := ReadForm(request)
		if err != nil {
			t.Error(err)
		}

		expectedForm := specs.Query{
			"hello": "world",
			"host":  "port",
			"key":   "value",
		}
		if !reflect.DeepEqual(form, expectedForm) {
			t.Errorf("read invalid form = %v, want %v", form, expectedForm)
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

	form := neturl.Values{}
	form.Set("hello", "world")
	form.Set("host", "port")
	form.Set("key", "value")
	resp, err := http.PostForm(url, form)
	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}
