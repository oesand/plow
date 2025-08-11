package ws

import (
	"bufio"
	"context"
	"fmt"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"slices"
	"strings"
)

// DefaultDialer returns a new Dialer with default settings.
func DefaultDialer() *Dialer {
	return &Dialer{}
}

// Dialer is used to create a WebSocket connection to a server.
type Dialer struct {
	_ internal.NoCopy

	// Origin is the value of the "Origin" header in the WebSocket handshake request.
	Origin string

	// Protocols is a list of subprotocols that the client supports.
	Protocols []string
}

// Dial creates a WebSocket connection to the specified URL using the provided client.
func (dialer *Dialer) Dial(client *giglet.Client, url specs.Url) (Conn, error) {
	ctx := context.Background()
	return dialer.dial(ctx, client, &url)
}

// DialContext creates a WebSocket connection to the specified URL using the provided client and context.
func (dialer *Dialer) DialContext(ctx context.Context, client *giglet.Client, url specs.Url) (Conn, error) {
	return dialer.dial(ctx, client, &url)
}

func (dialer *Dialer) dial(ctx context.Context, client *giglet.Client, url *specs.Url) (Conn, error) {
	if ctx == nil {
		panic("nil Context pointer")
	}
	if client == nil {
		panic("nil Client pointer")
	}
	if url.Host == "" {
		panic("empty url host")
	}

	if url.Scheme == "" {
		url.Scheme = "wss"
	}

	var httpScheme string
	switch url.Scheme {
	case "ws":
		httpScheme = "http"
	case "wss":
		httpScheme = "https"
	default:
		return nil, fmt.Errorf("unsupported scheme %s", url.Scheme)
	}

	httpUrl := *url
	httpUrl.Scheme = httpScheme

	hijacker, ctx := giglet.WithTransportHijacker(ctx)
	req := giglet.EmptyRequest(specs.HttpMethodGet, &httpUrl)

	req.Header().Set("Connection", "Upgrade")
	req.Header().Set("Upgrade", "websocket")
	req.Header().Set("Sec-Websocket-Version", "13")

	challengeKey := newChallengeKey()
	req.Header().Set("Sec-WebSocket-Key", string(challengeKey))

	if dialer.Origin != "" {
		req.Header().Set("Origin", dialer.Origin)
	}

	if len(dialer.Protocols) > 0 {
		req.Header().Set("Sec-WebSocket-Protocol", strings.Join(dialer.Protocols, ","))
	}

	resp, err := client.MakeContext(ctx, req)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode(); code != specs.StatusCodeSwitchingProtocols {
		return nil, specs.NewOpError("ws", "invalid status code %d", code)
	}

	if body := resp.Body(); body != nil {
		return nil, specs.NewOpError("ws", "response cannot have body")
	}

	if !strings.EqualFold(req.Header().Get("Connection"), "upgrade") ||
		!strings.EqualFold(req.Header().Get("Upgrade"), "websocket") {
		return nil, ErrFailChallenge
	}

	expectedAccept := computeAcceptKey(challengeKey)
	if resp.Header().Get("Sec-WebSocket-Accept") != expectedAccept {
		return nil, ErrFailChallenge
	}

	if resp.Header().Get("Sec-WebSocket-Extensions") != "" {
		return nil, specs.NewOpError("ws", "unsupported extensions")
	}

	var selectedProtocol string
	if len(dialer.Protocols) > 0 {
		if selectedProtocol = resp.Header().Get("Sec-WebSocket-Protocol"); !slices.Contains(dialer.Protocols, selectedProtocol) {
			return nil, ErrUnknownProtocol
		}
	}

	conn := hijacker.Conn
	if conn == nil {
		return nil, specs.NewOpError("ws", "fail to hijack connection")
	}

	reader := stream.DefaultBufioReaderPool.Get(conn)
	writer := stream.DefaultBufioWriterPool.Get(conn)
	rws := bufio.NewReadWriter(reader, writer)

	return newClientConn(*url, conn, rws, selectedProtocol), nil
}
