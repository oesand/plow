package ws

import (
	"bytes"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"testing"
	"time"
)

func newTestPipeConn(t *testing.T, compress bool) (server, client *wsConn) {
	t.Helper()
	s, c := net.Pipe()
	timeout := 2 * time.Second
	server = newWsConn(s, true, compress, 1024, timeout, timeout, "ws")
	client = newWsConn(c, false, compress, 1024, timeout, timeout, "ws")
	return
}

func TestWsConn_AliveAndClose(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	conn := newWsConn(server, true, false, 1024, time.Second, time.Second, "test")
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

// TODO : Add tests with built-in websocket package

/*


func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	t.Logf("Listening on %s", listener.Addr().String())

	wsServer := websocket.Server{}
	wsServer.Handler = func(conn *websocket.Conn) {
		for {
			var buf = make([]byte, 4)
			_, err := io.ReadFull(conn, buf)
			//buf, err := io.ReadAll(conn)
			if err != nil {
				t.Error(err)
				break
			}
			t.Logf("Server received: %s \n", buf)
			//fmt.Fprintf(conn, "Answer: %s", buf)
			conn.Write(buf)
		}
	}

	go func() {
		err = http.Serve(listener, wsServer)
		if err != nil {
			t.Fatal(err)
		}
	}()

	client := giglet.DefaultClient()

	conn, err := DefaultDialer().Dial(client, *specs.MustParseUrl("ws://" + listener.Addr().String()))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connected to %s", conn.RemoteAddr().String())

	var i int
	var input string = "000"
	for conn.Alive() {
		input += strconv.Itoa(i)
		_, err = conn.Write([]byte(input))
		if err != nil {
			t.Error(err)
			break
		}
		i++

		t.Logf("Sent: %s \n", input)

		//buf := make([]byte, len(input))
		buf, err := io.ReadAll(conn)
		//_, err = conn.Read(buf)
		if err != nil && err != io.EOF {
			t.Error(err)
			break
		}
		t.Logf("Client received: %s \n", buf)

		time.Sleep(1 * time.Second)
	}
}



func TestWs(t *testing.T) {
	//conn, err := websocket.Dial("wss://ws.postman-echo.com/raw", "", "http://localhost/")
	//if err != nil {
	//	t.Fatal(err)
	//}

	url := specs.MustParseUrl("wss://ws.postman-echo.com/raw")
	conn, err := DefaultDialer().Dial(giglet.DefaultClient(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	t.Logf("Connected to %s", conn.RemoteAddr().String())

	fmt.Printf("Compression: %v \n", conn.(*wsConn).compressEnabled)

	var i int
	var input string = "000"
	for {
		input += strconv.Itoa(i)
		_, err = conn.WriteText(input)
		if err != nil {
			t.Error(err)
			break
		}
		i++

		t.Logf("Sent: %s \n", input)

		//buf := make([]byte, len(input))
		buf, err := io.ReadAll(conn)
		//_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("Received: %s \n", buf)

		time.Sleep(1 * time.Second)
	}
}


*/
