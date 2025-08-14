package ws

import (
	"bytes"
	"context"
	"errors"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/mock"
	"github.com/oesand/giglet/specs"
	"io"
	"net"
	"strings"
	"testing"
)

func TestDialer_Panics(t *testing.T) {
	d := DefaultDialer()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil ctx")
		}
	}()
	d.DialContext(nil, &giglet.Client{}, &specs.Url{Host: "x"})
}

func TestDialer_UnsupportedScheme(t *testing.T) {
	d := DefaultDialer()
	_, err := d.DialContext(context.Background(), &giglet.Client{}, &specs.Url{Host: "x", Scheme: "ftp"})
	if err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestDialer_StatusCodeNotSwitching(t *testing.T) {
	d := DefaultDialer()
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			return mock.ClientResponse(specs.StatusCodeOK, nil), nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if err == nil || !strings.Contains(err.Error(), "invalid status code") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDialer_ResponseHasBody(t *testing.T) {
	d := DefaultDialer()
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			return mock.ClientResponse(specs.StatusCodeSwitchingProtocols, io.NopCloser(bytes.NewReader([]byte("x")))), nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if err == nil || !strings.Contains(err.Error(), "response cannot have body") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDialer_BadUpgradeHeaders(t *testing.T) {
	d := DefaultDialer()
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			resp := mock.ClientResponse(specs.StatusCodeSwitchingProtocols, nil)
			resp.Header().Set("Connection", "keep-alive")
			return resp, nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if !errors.Is(err, ErrFailChallenge) {
		t.Fatalf("expected ErrFailChallenge, got %v", err)
	}
}

func TestDialer_BadAcceptKey(t *testing.T) {
	d := DefaultDialer()
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			resp := mock.ClientResponse(specs.StatusCodeSwitchingProtocols, nil)
			resp.Header().Set("Connection", "Upgrade")
			resp.Header().Set("Upgrade", "websocket")
			resp.Header().Set("Sec-WebSocket-Accept", "wrong")
			return resp, nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if !errors.Is(err, ErrFailChallenge) {
		t.Fatalf("expected ErrFailChallenge, got %v", err)
	}
}

func TestDialer_ProtocolMismatch(t *testing.T) {
	d := DefaultDialer()
	d.Protocols = []string{"proto1"}
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			resp := mock.ClientResponse(specs.StatusCodeSwitchingProtocols, nil)
			resp.Header().Set("Connection", "Upgrade")
			resp.Header().Set("Upgrade", "websocket")
			resp.Header().Set("Sec-WebSocket-Accept", computeAcceptKey(header.Get("Sec-WebSocket-Key")))
			resp.Header().Set("Sec-WebSocket-Protocol", "unknown")
			return resp, nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if !errors.Is(err, ErrUnknownProtocol) {
		t.Fatalf("expected ErrUnknownProtocol, got %v", err)
	}
}

func TestDialer_NoHijackedConn(t *testing.T) {
	d := DefaultDialer()
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			resp := mock.ClientResponse(specs.StatusCodeSwitchingProtocols, nil)
			resp.Header().Set("Connection", "Upgrade")
			resp.Header().Set("Upgrade", "websocket")
			resp.Header().Set("Sec-WebSocket-Accept", computeAcceptKey(header.Get("Sec-WebSocket-Key")))
			return resp, nil
		}),
	}
	_, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if err == nil || !strings.Contains(err.Error(), "fail to hijack") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDialer_Success(t *testing.T) {
	d := DefaultDialer()
	d.Protocols = []string{"proto1"}
	c := &giglet.Client{
		Transport: giglet.RoundTripperFunc(func(ctx context.Context, method specs.HttpMethod, url specs.Url, header *specs.Header, writer giglet.BodyWriter) (giglet.ClientResponse, error) {
			resp := mock.ClientResponse(specs.StatusCodeSwitchingProtocols, nil)
			resp.Header().Set("Connection", "Upgrade")
			resp.Header().Set("Upgrade", "websocket")
			resp.Header().Set("Sec-WebSocket-Accept", computeAcceptKey(header.Get("Sec-WebSocket-Key")))
			resp.Header().Set("Sec-WebSocket-Protocol", "proto1")
			resp.Header().Set("Sec-WebSocket-Extensions", "permessage-deflate")
			hj, _ := giglet.WithTransportHijacker(ctx)
			conn, _ := net.Pipe()
			hj.Conn = conn
			return resp, nil
		}),
	}
	conn, err := d.DialContext(context.Background(), c, &specs.Url{Host: "x"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	wsconn := conn.(*wsConn)

	if wsconn.protocol != "proto1" {
		t.Fatal("incorrect ws protocol in connection")
	}

	if !wsconn.compressEnabled {
		t.Fatal("compression should be enabled in connection")
	}
}
