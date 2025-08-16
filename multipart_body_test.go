package plow

import (
	"bytes"
	"context"
	"github.com/oesand/plow/specs"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPostMultipartRequestChunked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != "chunked" {
			t.Errorf("invalid transfer encoding %+v", r.TransferEncoding)
		}

		err := r.ParseMultipartForm(1024 * 1024) // 1MB limit
		if err != nil {
			t.Error(err)
		}
		form := r.MultipartForm

		if !reflect.DeepEqual(form.Value["hello"], []string{"world"}) ||
			!reflect.DeepEqual(form.Value["host"], []string{"port"}) ||
			!reflect.DeepEqual(form.Value["key"], []string{"value"}) {

			t.Errorf("unexpected form, got %v", form)
		}

		w.Write([]byte("received"))
	}))
	defer server.Close()

	url := specs.MustParseUrl(server.URL)
	req := MultipartRequest(specs.HttpMethodPost, url, func(w *multipart.Writer) error {
		err := w.WriteField("hello", "world")
		if err != nil {
			return err
		}
		err = w.WriteField("host", "port")
		if err != nil {
			return err
		}
		return w.WriteField("key", "value")
	})

	resp, err := DefaultClient().Make(req)
	if err != nil {
		t.Fatal("req:", err)
	}

	checkResponseBody(t, resp, []byte("received"))
}

func TestMultipartReader(t *testing.T) {
	server := DefaultServer(HandlerFunc(func(ctx context.Context, request Request) Response {
		reader, err := MultipartReader(request)
		if err != nil {
			t.Error(err)
		}
		form, err := reader.ReadForm(1024 * 1024) // 1MB limit
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(form.Value["hello"], []string{"world"}) ||
			!reflect.DeepEqual(form.Value["host"], []string{"port"}) ||
			!reflect.DeepEqual(form.Value["key"], []string{"value"}) {

			t.Errorf("unexpected form, got %v", form)
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

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("hello", "world")
	writer.WriteField("host", "port")
	writer.WriteField("key", "value")
	writer.Close()

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		t.Fatal("req:", err)
	}

	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("not found expected headers, %+v", resp.Header)
	}

	checkHttpResponseBody(t, resp, []byte("okay"))
}
