package giglet

import (
	"bytes"
	"context"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"testing"
	"time"
)

func checkResponseBody(t *testing.T, resp ClientResponse, expected []byte) {
	if resp.StatusCode() != specs.StatusCodeOK {
		t.Fatal("invalid status code:", resp.StatusCode())
	}

	body := resp.Body()
	if body == nil {
		t.Fatal("response body is nil")
	}

	defer body.Close()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatal("read all:", err)
	}

	if !bytes.Equal(data, expected) {
		t.Error("invalid response:", string(data))
	}
}

func newTestServer(handler func(header *specs.Header) (specs.StatusCode, []byte)) func() {
	listener, err := net.Listen("tcp4", ":http")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

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
			}
			break

			select {
			case <-ctx.Done():
			default:
			}
		}

		header := specs.NewHeader()
		code, body := handler(header)
		server.WriteResponseHead(conn, true, code, header)
		conn.Write(body)
		conn.Close()
	}()

	closeFunc := func() {
		listener.Close()
		cancel()
	}

	return closeFunc
}
