package plow

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/internal/client_ops"
	"github.com/oesand/plow/internal/server_ops"
	"github.com/oesand/plow/specs"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func checkResponseBody(t *testing.T, resp ClientResponse, expected []byte) {
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

	if !bytes.Equal(data, expected) {
		t.Error("invalid response:", string(data))
	}
}

func checkHttpResponseBody(t *testing.T, resp *http.Response, expected []byte) {
	if resp.StatusCode != 200 {
		t.Fatal("invalid status code:", resp.StatusCode)
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

	if !bytes.Equal(data, expected) {
		t.Error("invalid response:", string(data))
	}
}

func serveTcpTest(ctx context.Context, handler func(net.Conn)) (*specs.Url, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	go func() {
		for {
			conn, err := listener.Accept()

			if ctx.Err() != nil {
				if err == nil {
					conn.Close()
				}
				listener.Close()
				return
			}

			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
			}

			handler(conn)

			listener.Close()
			break
		}
	}()

	return url, err
}

func newTestServer(handler func(req Request) (specs.StatusCode, *specs.Header, []byte)) (func(), *specs.Url) {
	ctx, cancel := context.WithCancel(context.Background())

	url, err := serveTcpTest(ctx, func(conn net.Conn) {
		bufioReader := bufio.NewReader(conn)

		req, err := server_ops.ReadRequest(ctx, conn.RemoteAddr(), bufioReader, 1024, 8*1024)
		if err != nil {
			panic(err)
		}
		req.BodyReader = bufioReader

		code, header, body := handler(req)
		if header == nil {
			header = specs.NewHeader()
		}
		server_ops.WriteResponseHead(conn, true, code, header)
		if body != nil {
			conn.Write(body)
		}
		if hijacker := req.Hijacker(); hijacker != nil {
			hijacker(ctx, conn)
		}

		conn.Close()
	})
	if err != nil {
		panic(err)
	}

	return cancel, url
}

func newTestClientSend(method specs.HttpMethod, url *specs.Url, header *specs.Header, body []byte) (ClientResponse, net.Conn, error) {
	address := client_ops.HostPort(url.Host, url.Port)
	conn, err := defaultDialer.Dial("tcp", address)
	if err != nil {
		return nil, nil, err
	}
	conn.SetDeadline(time.Now().Add(20 * time.Second))

	_, err = client_ops.WriteRequestHead(conn, method, url.Path, url.Query, header)
	if err != nil {
		return nil, nil, err
	}

	if body != nil {
		_, err = conn.Write(body)
		if err != nil {
			return nil, nil, err
		}
	}

	headerReader := bufio.NewReader(conn)

	resp, err := client_ops.ReadResponse(context.Background(), headerReader, 1024, 8*1024)
	if err != nil {
		return nil, nil, err
	}

	extraBuffered, _ := headerReader.Peek(headerReader.Buffered())

	resp.Reader = internal.ReadCloser(io.MultiReader(
		bytes.NewReader(extraBuffered), conn), conn)

	return resp, conn, nil
}
