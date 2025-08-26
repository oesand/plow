package ws

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func newTestPipeConn(t *testing.T, compress bool) (server, client *wsConn) {
	t.Helper()
	s, c := net.Pipe()
	timeout := 2 * time.Second
	ctx := context.Background()
	server = newWsConn(ctx, s, true, compress, 1024, timeout, timeout, "ws")
	client = newWsConn(ctx, c, false, compress, 1024, timeout, timeout, "ws")
	return
}

func TestWsConn_AliveAndClose(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	conn := newWsConn(context.Background(), server, true, false, 1024, time.Second, time.Second, "test")
	if !conn.Alive() {
		t.Fatal("conn should be alive initially")
	}

	err := conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	if conn.Alive() {
		t.Fatal("conn should be dead after Close")
	}

	err = conn.Close()
	if err != specs.ErrClosed {
		t.Fatalf("expected ErrClosed on double close, got %v", err)
	}
}

func TestServer_ClientClose(t *testing.T) {
	server, client := newTestPipeConn(t, false)
	defer server.Close()
	defer client.Close()

	go func() {
		err := client.WriteClose(CloseCodeNormal)
		if err != nil {
			t.Error(err)
		}
	}()

	buf := make([]byte, 1024)
	_, err := server.Read(buf)
	if err != nil && err != specs.ErrClosed {
		t.Fatal(err)
	}
}

// Test Client send and Server receive

func TestServer_BinaryFrame(t *testing.T) {
	server, client := newTestPipeConn(t, false)
	defer server.Close()
	defer client.Close()

	msg := []byte("hello binary")

	// Client writes to server
	go func() {
		n, err := client.Write(msg)
		if err != nil {
			t.Error(err)
		}
		if n != len(msg) {
			t.Errorf("expected %d written, got %d", len(msg), n)
		}
	}()

	buf, err := io.ReadAll(server)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, msg) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestServer_TextFrame(t *testing.T) {
	server, client := newTestPipeConn(t, false)
	defer server.Close()
	defer client.Close()

	msg := "hello text"

	go func() {
		n, err := client.WriteText(msg)
		if err != nil {
			t.Error(err)
		}
		if n != len(msg) {
			t.Errorf("expected %d written, got %d", len(msg), n)
		}
	}()

	buf, err := io.ReadAll(server)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte(msg)) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestServer_CompressBinaryFrame(t *testing.T) {
	server, client := newTestPipeConn(t, true)
	defer server.Close()
	defer client.Close()

	msg := []byte("compressed message content")

	go func() {
		n, err := client.Write(msg)
		if err != nil {
			t.Error(err)
		}
		if n == 0 {
			t.Error("expected bytes written")
		}
	}()

	buf, err := io.ReadAll(server)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, msg) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestServer_CompressTextFrame(t *testing.T) {
	server, client := newTestPipeConn(t, true)
	defer server.Close()
	defer client.Close()

	msg := "hello text"

	go func() {
		n, err := client.WriteText(msg)
		if err != nil {
			t.Error(err)
		}
		if n == 0 {
			t.Error("expected bytes written")
		}
	}()

	buf, err := io.ReadAll(server)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte(msg)) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

// Test Server send and Client receive

func TestClient_BinaryFrame(t *testing.T) {
	server, client := newTestPipeConn(t, false)
	defer server.Close()
	defer client.Close()

	msg := []byte("hello binary")

	// Client writes to server
	go func() {
		n, err := server.Write(msg)
		if err != nil {
			t.Error(err)
		}
		if n != len(msg) {
			t.Errorf("expected %d written, got %d", len(msg), n)
		}
	}()

	buf, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, msg) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestClient_TextFrame(t *testing.T) {
	server, client := newTestPipeConn(t, false)
	defer server.Close()
	defer client.Close()

	msg := "hello text"

	go func() {
		n, err := server.WriteText(msg)
		if err != nil {
			t.Error(err)
		}
		if n != len(msg) {
			t.Errorf("expected %d written, got %d", len(msg), n)
		}
	}()

	buf, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte(msg)) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestClient_CompressBinaryFrame(t *testing.T) {
	server, client := newTestPipeConn(t, true)
	defer server.Close()
	defer client.Close()

	msg := []byte("hello binary")

	// Client writes to server
	go func() {
		n, err := server.Write(msg)
		if err != nil {
			t.Error(err)
		}
		if n == 0 {
			t.Error("expected bytes written")
		}
	}()

	buf, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, msg) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

func TestClient_CompressTextFrame(t *testing.T) {
	server, client := newTestPipeConn(t, true)
	defer server.Close()
	defer client.Close()

	msg := "hello text"

	go func() {
		n, err := server.WriteText(msg)
		if err != nil {
			t.Error(err)
		}
		if n == 0 {
			t.Error("expected bytes written")
		}
	}()

	buf, err := io.ReadAll(client)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, []byte(msg)) {
		t.Errorf("expected %s got %s", msg, buf)
	}
}

// Test conn with net/websocket package

func TestDialNetWebsocket(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	t.Logf("Listening on %s", listener.Addr().String())

	wsServer := websocket.Server{}
	wsServer.Handler = func(conn *websocket.Conn) {
		var input = "000"
		var i int
		for {
			input += strconv.Itoa(i)
			var buf = make([]byte, len(input))
			_, err := io.ReadFull(conn, buf)
			if err != nil {
				t.Error(err)
				break
			}
			if !bytes.Equal(buf, []byte(input)) {
				t.Fatalf("Invalid server received: %s \n", buf)
			}
			i++
			t.Logf("Server received: %s \n", buf)
			fmt.Fprintf(conn, "Answer: %s", buf)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		err = http.Serve(listener, wsServer)
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
		if err != nil {
			t.Error(err)
		}
	}()

	client := plow.DefaultClient()
	dialer := DefaultDialer()
	dialer.EnableCompression = false
	conn, err := dialer.DialContext(ctx, client, specs.MustParseUrl("ws://"+listener.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connected to %s", conn.RemoteAddr().String())

	var i int
	var input = "000"
	for conn.Alive() {
		if i >= 3 {
			break
		}

		input += strconv.Itoa(i)
		_, err = conn.Write([]byte(input))
		if err != nil {
			t.Error(err)
			break
		}

		t.Logf("Sent: %s \n", input)

		buf, err := io.ReadAll(conn)
		if err != nil {
			t.Error(err)
			break
		}
		if !bytes.Equal(buf, []byte("Answer: "+input)) {
			t.Fatalf("Invalid received: %s \n", buf)
		}
		t.Logf("Client received: %s \n", buf)
		i++

		time.Sleep(50 * time.Millisecond)
	}
}

func TestUpgraderNetWebsocket(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	t.Logf("Listening on %s", listener.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var input = "000"
	upgrader := DefaultUpgrader()
	upgrader.EnableCompression = false
	wsServer := plow.DefaultServer(plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		return upgrader.Upgrade(request, func(ctx context.Context, conn Conn) {
			t.Logf("Received conn %s", conn.RemoteAddr().String())
			for conn.Alive() {
				buf, err := io.ReadAll(conn)
				if err != nil {
					t.Error(err)
					break
				}
				if !bytes.Equal(buf, []byte(input)) {
					t.Fatalf("Invalid server received: %s \n", buf)
				}
				t.Logf("Server received: %s \n", buf)
				fmt.Fprintf(conn, "Answer: %s", buf)

				time.Sleep(50 * time.Millisecond)
			}
			t.Logf("Dead server conn")
			defer cancel()
		})
	}))

	go func() {
		err = wsServer.Serve(listener)
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		}
		if err != nil {
			t.Error(err)
		}
	}()

	conn, err := websocket.Dial("ws://"+listener.Addr().String(), "", "http://127.0.0.1/")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connected to %s", conn.RemoteAddr().String())

	var i int32
	for {
		if i >= 5 {
			break
		}

		input += strconv.Itoa(int(i))
		_, err = conn.Write([]byte(input))
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("Sent client: %s\n", input)
		i++

		var buf = make([]byte, len(input)+8)
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			t.Error(err)
			break
		}
		if !bytes.Equal(buf, []byte("Answer: "+input)) {
			t.Errorf("Invalid client received: %s \n", buf)
		}
		t.Logf("Client received: %s \n", buf)
	}
}
