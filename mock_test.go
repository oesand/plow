package giglet

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"testing"
)

func checkResponseBody(t *testing.T, resp ClientResponse, expected []byte) {
	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode(), ", body:", string(data))
	}

	if !bytes.Equal(data, expected) {
		t.Error("invalid response:", string(data))
	}
}

func newTestServer(handler Handler) (close func()) {
	server := DefaultServer(handler)

	listener, err := net.Listen("tcp4", ":http")
	if err != nil {
		panic(err)
	}

	go server.Serve(listener)

	return func() {
		go server.Shutdown()
		listener.Close()
	}
}
