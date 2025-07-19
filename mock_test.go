package giglet

import (
	"bufio"
	"bytes"
	"context"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/server"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"net/http"
	"strconv"
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

func newTestServer(handler func(header *specs.Header) (specs.StatusCode, []byte)) (func(), *specs.Url) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
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

	url := specs.MustParseUrl("http://" + listener.Addr().String())

	return closeFunc, url
}

func newTestClientSend(method specs.HttpMethod, url *specs.Url, header *specs.Header, body []byte) (ClientResponse, net.Conn, error) {
	address := url.Host + ":" + strconv.FormatUint(uint64(url.Port), 10)
	conn, err := defaultDialer.Dial("tcp", address)
	if err != nil {
		return nil, nil, err
	}
	conn.SetDeadline(time.Now().Add(20 * time.Second))

	_, err = client.WriteRequestHead(conn, method, url, header)
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

	resp, err := client.ReadResponse(context.Background(), headerReader, 1024, 8*1024)
	if err != nil {
		return nil, nil, err
	}

	extraBuffered, _ := headerReader.Peek(headerReader.Buffered())

	resp.Reader = internal.ReadCloser(io.MultiReader(
		bytes.NewReader(extraBuffered), conn), conn)

	return resp, conn, nil
}
