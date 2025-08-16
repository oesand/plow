package ws

import (
	"context"
	"github.com/oesand/plow"
	"github.com/oesand/plow/mock"
	"github.com/oesand/plow/specs"
	"testing"
)

func TestUpgrader_Upgrade_FailureCases(t *testing.T) {
	cases := []struct {
		name       string
		req        plow.Request
		wantStatus specs.StatusCode
	}{
		{
			name: "invalid method",
			req: mock.DefaultRequest().
				Method(specs.HttpMethodPost).
				Request(),
			wantStatus: specs.StatusCodeMethodNotAllowed,
		},
		{
			name: "missing connection header",
			req: mock.DefaultRequest().
				Method(specs.HttpMethodGet).
				Request(),
			wantStatus: specs.StatusCodeBadRequest,
		},
		{
			name: "invalid upgrade header",
			req: mock.DefaultRequest().
				Method(specs.HttpMethodGet).
				ConfHeader(func(header *specs.Header) {
					header.Set("Connection", "Upgrade")
					header.Set("Upgrade", "notwebsocket")
				}).
				Request(),
			wantStatus: specs.StatusCodeBadRequest,
		},
		{
			name: "unsupported websocket version",
			req: mock.DefaultRequest().
				Method(specs.HttpMethodGet).
				ConfHeader(func(header *specs.Header) {
					header.Set("Connection", "Upgrade")
					header.Set("Upgrade", "websocket")
					header.Set("Sec-Websocket-Version", "12")
				}).
				Request(),
			wantStatus: specs.StatusCodeNotImplemented,
		},
		{
			name: "missing Sec-Websocket-Key",
			req: mock.DefaultRequest().
				Method(specs.HttpMethodGet).
				ConfHeader(func(header *specs.Header) {
					header.Set("Connection", "Upgrade")
					header.Set("Upgrade", "websocket")
					header.Set("Sec-Websocket-Version", "13")
				}).
				Request(),
			wantStatus: specs.StatusCodeBadRequest,
		},
	}

	upgrader := DefaultUpgrader()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := upgrader.Upgrade(c.req, func(ctx context.Context, ws Conn) {})
			if resp.StatusCode() != c.wantStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode(), c.wantStatus)
			}
		})
	}
}

func TestUpgrader_Upgrade_Success(t *testing.T) {
	challengeKey, err := newChallengeKey()
	if err != nil {
		t.Fatalf("failed to generate challenge key: %v", err)
	}

	builder := mock.DefaultRequest().
		Method(specs.HttpMethodGet).
		ConfHeader(func(header *specs.Header) {
			header.Set("Connection", "Upgrade")
			header.Set("Upgrade", "websocket")
			header.Set("Sec-Websocket-Version", "13")
			header.Set("Sec-Websocket-Key", challengeKey)
		})

	req := builder.Request()

	upgrader := DefaultUpgrader()
	resp := upgrader.Upgrade(req, func(ctx context.Context, ws Conn) {})

	if resp.StatusCode() != specs.StatusCodeSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", specs.StatusCodeSwitchingProtocols, resp.StatusCode())
	}
	if builder.Hijacker() == nil {
		t.Fatal("expected hijack to be called")
	}
	if resp.Header().Get("Upgrade") != "websocket" {
		t.Error("missing Upgrade header")
	}
	if resp.Header().Get("Connection") != "Upgrade" {
		t.Error("missing Connection header")
	}
	if resp.Header().Get("Sec-WebSocket-Accept") != computeAcceptKey(challengeKey) {
		t.Error("incorrect Sec-WebSocket-Accept header")
	}
}

// --- Protocol selection ---
func TestUpgrader_Upgrade_ProtocolSelection(t *testing.T) {
	challengeKey, err := newChallengeKey()
	if err != nil {
		t.Fatalf("failed to generate challenge key: %v", err)
	}

	req := mock.DefaultRequest().
		Method(specs.HttpMethodGet).
		ConfHeader(func(header *specs.Header) {
			header.Set("Connection", "Upgrade")
			header.Set("Upgrade", "websocket")
			header.Set("Sec-Websocket-Version", "13")
			header.Set("Sec-Websocket-Key", challengeKey)
			header.Set("Sec-Websocket-Protocol", "a, b, c")
		}).
		Request()

	upgrader := &Upgrader{
		SelectProtocol:    func(protocols []string) string { return protocols[1] },
		EnableCompression: false,
	}

	resp := upgrader.Upgrade(req, func(ctx context.Context, conn Conn) {
		wsconn := conn.(*wsConn)

		if wsconn.protocol != "proto1" {
			t.Fatal("incorrect ws protocol in connection")
		}
	})

	if resp.Header().Get("Sec-WebSocket-Protocol") != "b" {
		t.Errorf("expected selected protocol 'b', got %q", resp.Header().Get("Sec-WebSocket-Protocol"))
	}
}

// --- Compression ---
func TestUpgrader_Upgrade_Compression(t *testing.T) {
	challengeKey, err := newChallengeKey()
	if err != nil {
		t.Fatalf("failed to generate challenge key: %v", err)
	}

	req := mock.DefaultRequest().
		Method(specs.HttpMethodGet).
		ConfHeader(func(header *specs.Header) {
			header.Set("Connection", "Upgrade")
			header.Set("Upgrade", "websocket")
			header.Set("Sec-Websocket-Version", "13")
			header.Set("Sec-Websocket-Key", challengeKey)
			header.Set("Sec-WebSocket-Extensions", "permessage-deflate")
		}).
		Request()

	upgrader := &Upgrader{
		EnableCompression: true,
	}

	resp := upgrader.Upgrade(req, func(ctx context.Context, conn Conn) {
		wsconn := conn.(*wsConn)

		if !wsconn.compressEnabled {
			t.Fatal("compression should be enabled in connection")
		}
	})

	if resp.Header().Get("Sec-WebSocket-Extensions") != "permessage-deflate" {
		t.Errorf("expected compression header, got %q", resp.Header().Get("Sec-WebSocket-Extensions"))
	}
}
