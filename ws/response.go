package ws

import (
	"bufio"
	"context"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"net"
	"strings"
)

func UpgradeResponse(req giglet.Request, handler Handler) giglet.Response {
	if req.Method() != specs.HttpMethodGet {
		return giglet.TextResponse(specs.StatusCodeMethodNotAllowed, specs.ContentTypePlain,
			"websocket: upgrading required request method - GET")
	} else if !strings.EqualFold(req.Header().Get("Connection"), "upgrade") {
		return giglet.TextResponse(specs.StatusCodeBadRequest, specs.ContentTypePlain,
			"websocket: 'Upgrade' token not found in 'Connection' header")
	} else if !strings.EqualFold(req.Header().Get("Upgrade"), "websocket") {
		return giglet.TextResponse(specs.StatusCodeBadRequest, specs.ContentTypePlain,
			"websocket: 'websocket' token not found in 'Upgrade' header")
	} else if req.Header().Get("Sec-Websocket-Version") != "13" {
		return giglet.TextResponse(specs.StatusCodeNotImplemented, specs.ContentTypePlain,
			"websocket: supports only websocket 13 version")
	}

	challengeKey := req.Header().Get("Sec-Websocket-Key")
	if len(challengeKey) == 0 {
		return giglet.TextResponse(specs.StatusCodeBadRequest, specs.ContentTypePlain,
			"websocket: not a websocket handshake: `Sec-WebSocket-Key' header is missing or blank")
	}

	var challengeProtocols []string
	protocol := strings.TrimSpace(req.Header().Get("Sec-Websocket-Protocol"))
	if protocol != "" {
		protocols := strings.Split(protocol, ",")
		for i := 0; i < len(protocols); i++ {
			challengeProtocols = append(challengeProtocols, strings.TrimSpace(protocols[i]))
		}
	}

	req.Hijack(func(ctx context.Context, conn net.Conn) {
		reader := stream.DefaultBufioReaderPool.Get(conn)
		defer stream.DefaultBufioReaderPool.Put(reader)

		writer := stream.DefaultBufioWriterPool.Get(conn)
		defer stream.DefaultBufioWriterPool.Put(writer)

		rws := bufio.NewReadWriter(reader, writer)
		wsConn := newServerConn(req, conn, rws)

		handler(ctx, wsConn)
		wsConn.dead = true
	})

	return giglet.EmptyResponse(specs.StatusCodeSwitchingProtocols, func(resp giglet.Response) {
		resp.Header().Set("Upgrade", "websocket")
		resp.Header().Set("Connection", "Upgrade")
		resp.Header().Set("Sec-WebSocket-Accept", computeAcceptKey([]byte(challengeKey)))

		if len(challengeProtocols) > 0 {
			resp.Header().Set("Sec-WebSocket-Protocol", challengeProtocols[0])
		}
	})
}
