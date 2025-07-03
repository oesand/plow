package ws

import (
	"bufio"
	"context"
	"fmt"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"slices"
	"strings"
)

type DialConf struct {
	Origin    string
	Protocols []string
}

func Dial(client *giglet.Client, url specs.Url) (Conn, error) {
	ctx := context.Background()
	return dial(ctx, client, &url, &DialConf{})
}

func DialContext(ctx context.Context, client *giglet.Client, url specs.Url) (Conn, error) {
	return dial(ctx, client, &url, &DialConf{})
}

func DialContextConf(ctx context.Context, client *giglet.Client, url specs.Url, conf DialConf) (Conn, error) {
	return dial(ctx, client, &url, &conf)
}

func dial(ctx context.Context, client *giglet.Client, url *specs.Url, conf *DialConf) (Conn, error) {
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
		panic(fmt.Sprintf("unsupported scheme %s", url.Scheme))
	}

	httpUrl := *url
	httpUrl.Scheme = httpScheme

	req := giglet.NewHijackRequest(specs.HttpMethodGet, &httpUrl)

	req.Header().Set("Connection", "Upgrade")
	req.Header().Set("Upgrade", "websocket")
	req.Header().Set("Sec-Websocket-Version", "13")

	challengeKey := newChallengeKey()
	req.Header().Set("Sec-WebSocket-Key", string(challengeKey))

	if conf.Origin != "" {
		req.Header().Set("Origin", conf.Origin)
	}

	if len(conf.Protocols) > 0 {
		req.Header().Set("Sec-WebSocket-Protocol", strings.Join(conf.Protocols, ","))
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

	if len(conf.Protocols) > 0 {
		if proto := resp.Header().Get("Sec-WebSocket-Protocol"); !slices.Contains(conf.Protocols, proto) {
			return nil, ErrUnknownProtocol
		}
	}

	conn := req.Conn()
	if conn == nil {
		return nil, specs.NewOpError("ws", "fail to hijack connection")
	}

	reader := stream.DefaultBufioReaderPool.Get(conn)
	writer := stream.DefaultBufioWriterPool.Get(conn)
	rws := bufio.NewReadWriter(reader, writer)

	return newClientConn(*url, conn, rws), nil
}
