package proxy

import (
	"bytes"
	"github.com/armon/go-socks5"
	"net"
	"strconv"
	"testing"
	"time"
)

func newTestTcpServer(handler func(conn net.Conn)) (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go func() {
		var conn net.Conn
		for {
			conn, err = listener.Accept()

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					time.Sleep(5 * time.Millisecond)
					continue
				}
			}
			break
		}
		handler(conn)
	}()
	return listener, nil
}

func splitHostPort(addr string) (string, uint16) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}

	num, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		panic(err)
	}
	return host, uint16(num)
}

func TestDialSocks5_AllAuthMethods(t *testing.T) {
	tests := []struct {
		name  string
		creds *Creds
	}{
		{name: "No Auth"},
		{
			name:  "No Password",
			creds: &Creds{Username: "username"},
		},
		{
			name:  "With Password",
			creds: &Creds{"username", "password"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyConf := &socks5.Config{}
			if tt.creds != nil {
				proxyConf.Credentials = &socks5.StaticCredentials{
					tt.creds.Username: tt.creds.Password,
				}
			}
			server, err := socks5.New(proxyConf)
			if err != nil {
				t.Fatal(err)
			}

			proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}

			go func() {
				if err := server.Serve(proxyListener); err != nil {
					t.Error(err)
				}
			}()

			listener, err := newTestTcpServer(func(conn net.Conn) {
				buf := make([]byte, 4)
				_, err := conn.Read(buf)
				if err != nil {
					t.Error(err)
				}
				if !bytes.Equal(buf, []byte("ping")) {
					t.Error("invalid ping message")
				}
				_, err = conn.Write([]byte("pong"))
				if err != nil {
					t.Error(err)
				}
			})

			conn, err := net.Dial("tcp", proxyListener.Addr().String())
			if err != nil {
				t.Error(err)
			}

			host, port := splitHostPort(listener.Addr().String())
			_, err = DialSocks5(conn, host, port, tt.creds)
			if err != nil {
				t.Fatal(err)
			}

			// Handle Response
			_, err = conn.Write([]byte("ping"))
			if err != nil {
				t.Error(err)
			}

			buf := make([]byte, 4)
			_, err = conn.Read(buf)
			if err != nil {
				t.Error(err)
			}
			if !bytes.Equal(buf, []byte("pong")) {
				t.Error("invalid pong message")
			}
		})
	}
}

func TestDialSocks5_AuthErrors(t *testing.T) {
	tests := []struct {
		name    string
		creds   *Creds
		wantErr string
	}{
		{
			name:    "Require Auth",
			wantErr: "socks5: no acceptable authentication methods",
		},
		{
			name:    "Unknown Username",
			wantErr: "socks5: username/password authentication failed",
			creds:   &Creds{Username: "invalid"},
		},
		{
			name:    "Invalid Password",
			wantErr: "socks5: username/password authentication failed",
			creds:   &Creds{"username", "invalid"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := socks5.New(&socks5.Config{
				Credentials: &socks5.StaticCredentials{
					"username": "password",
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}

			go func() {
				if err := server.Serve(proxyListener); err != nil {
					t.Error(err)
				}
			}()

			conn, err := net.Dial("tcp", proxyListener.Addr().String())
			if err != nil {
				t.Error(err)
			}

			host, port := splitHostPort("127.0.0.1:1243")
			_, err = DialSocks5(conn, host, port, tt.creds)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("got unexpected error '%v', expected '%v'", err, tt.wantErr)
			}
		})
	}
}
