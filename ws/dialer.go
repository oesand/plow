package ws

import (
	"context"
	"fmt"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"slices"
	"strings"
	"time"
)

// DefaultDialer returns a new Dialer with default settings.
func DefaultDialer() *Dialer {
	return &Dialer{
		EnableCompression: true,
		MaxFrameSize:      8 * 1024, // 8KB
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
}

// Dialer is used to create a WebSocket connection to a server.
type Dialer struct {
	_ internal.NoCopy

	// Origin is the value of the "Origin" header in the WebSocket handshake request.
	Origin string

	// Protocols is a list of subprotocols that the client supports.
	Protocols []string

	// EnableCompression indicates whether to enable 'permessage-deflate' extension for compression.
	EnableCompression bool

	// MaxFrameSize is the maximum size of a WebSocket frame in bytes.
	MaxFrameSize int

	// ReadTimeout indicates the maximum duration for reading messages from the WebSocket connection.
	ReadTimeout time.Duration

	// WriteTimeout indicates the maximum duration for writing messages to the WebSocket connection.
	WriteTimeout time.Duration
}

// Dial creates a WebSocket connection to the specified URL using the provided client.
func (dialer *Dialer) Dial(client *giglet.Client, url specs.Url, configure ...func(giglet.ClientRequest)) (Conn, error) {
	ctx := context.Background()
	return dialer.dial(ctx, client, &url, configure...)
}

// DialContext creates a WebSocket connection to the specified URL using the provided client and context.
func (dialer *Dialer) DialContext(ctx context.Context, client *giglet.Client, url specs.Url, configure ...func(giglet.ClientRequest)) (Conn, error) {
	return dialer.dial(ctx, client, &url, configure...)
}

func (dialer *Dialer) dial(ctx context.Context, client *giglet.Client, url *specs.Url, configure ...func(giglet.ClientRequest)) (Conn, error) {
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

	for _, conf := range configure {
		conf(req)
	}

	req.Header().Set("Connection", "Upgrade")
	req.Header().Set("Upgrade", "websocket")
	req.Header().Set("Sec-Websocket-Version", "13")

	challengeKey := newChallengeKey()
	req.Header().Set("Sec-WebSocket-Key", string(challengeKey))

	if dialer.Origin != "" {
		req.Header().Set("Origin", dialer.Origin)
	}

	if dialer.EnableCompression {
		req.Header().Set("Sec-WebSocket-Extensions", "permessage-deflate")
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

	var compression bool
	if dialer.EnableCompression {
		extensions := resp.Header().Get("Sec-WebSocket-Extensions")
		compression = strings.Contains(extensions, "permessage-deflate")
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

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	wsConn := newWsConn(
		conn,
		false,
		compression,

		dialer.MaxFrameSize,
		dialer.ReadTimeout,
		dialer.WriteTimeout,

		selectedProtocol,
	)

	return wsConn, nil
}
