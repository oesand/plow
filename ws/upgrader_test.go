package ws

/*

func TestJustTest(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	server := giglet.DefaultServer(giglet.HandlerFunc(func(ctx context.Context, request giglet.Request) giglet.Response {
		return DefaultUpgrader().Upgrade(request, func(ctx context.Context, conn Conn) {
			for conn.Alive() {
				buf, err := io.ReadAll(conn)
				if err != nil {
					t.Error(err)
					break
				}
				t.Logf("Received: %s \n", buf)
				fmt.Fprintf(conn, "Answer: %s", buf)
			}
		})
	}))

	t.Logf("Listening on %s", listener.Addr().String())
	err = server.ServeTLSRaw(listener, mock.NewTlsCert())
	if err != nil {
		t.Fatal(err)
	}
}

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
			//var buf = make([]byte, 4)
			//_, err := io.ReadFull(conn, buf)
			buf, err := io.ReadAll(conn)
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
	conn, err := websocket.Dial("wss://ws.postman-echo.com/raw", "", "http://localhost/")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Connected to %s", conn.RemoteAddr().String())

	var i int
	var input string = "000"
	for {
		input += strconv.Itoa(i)
		_, err = conn.Write([]byte(input))
		if err != nil {
			t.Error(err)
			break
		}
		i++

		t.Logf("Sent: %s \n", input)

		buf := make([]byte, len(input))
		//buf, err := io.ReadAll(conn)
		_, err = conn.Read(buf)
		if err != nil {
			t.Error(err)
			break
		}
		t.Logf("Received: %s \n", buf)

		time.Sleep(1 * time.Second)
	}
}


*/
