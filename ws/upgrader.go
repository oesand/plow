package ws

import (
	"bufio"
	"context"
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"net"
	"strings"
)

// DefaultUpgrader returns a default Upgrader instance with no custom protocol selection.
func DefaultUpgrader() *Upgrader {
	return &Upgrader{}
}

// Upgrader is used to upgrade HTTP requests to WebSocket connections.
type Upgrader struct {
	_ internal.NoCopy

	// SelectProtocol is an optional function to select a protocol from the
	// Sec-WebSocket-Protocol header. If nil, the first protocol from the header will
	// be selected by default.
	SelectProtocol func(protocols []string) string
}

// Upgrade upgrades an HTTP request to a WebSocket connection. It checks the request
// method, headers, and the Sec-WebSocket-Key. If the request is valid, it
// hijacks the connection and invokes the provided handler with a new WebSocket
// connection. The response is returned with the necessary headers to complete the
// WebSocket handshake.
func (upgrader *Upgrader) Upgrade(req giglet.Request, handler Handler) giglet.Response {
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
	if challengeKey == "" {
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

	var selectedProtocol string
	if len(challengeProtocols) > 0 {
		if upgrader.SelectProtocol != nil {
			selectedProtocol = upgrader.SelectProtocol(challengeProtocols)
			if selectedProtocol == "" {
				return giglet.TextResponse(specs.StatusCodeNotImplemented, specs.ContentTypePlain,
					"websocket: not found supported protocols from `Sec-WebSocket-Protocol` header")
			}
		} else {
			selectedProtocol = challengeProtocols[0]
		}
	}

	req.Hijack(func(ctx context.Context, conn net.Conn) {
		reader := stream.DefaultBufioReaderPool.Get(conn)
		defer stream.DefaultBufioReaderPool.Put(reader)

		writer := stream.DefaultBufioWriterPool.Get(conn)
		defer stream.DefaultBufioWriterPool.Put(writer)

		rws := bufio.NewReadWriter(reader, writer)
		wsConn := newServerConn(req, conn, rws, selectedProtocol)

		handler(ctx, wsConn)
		wsConn.dead = true
	})

	acceptKey := computeAcceptKey([]byte(challengeKey))

	return giglet.EmptyResponse(specs.StatusCodeSwitchingProtocols, func(resp giglet.Response) {
		resp.Header().Set("Upgrade", "websocket")
		resp.Header().Set("Connection", "Upgrade")
		resp.Header().Set("Sec-WebSocket-Accept", acceptKey)

		if selectedProtocol != "" {
			resp.Header().Set("Sec-WebSocket-Protocol", selectedProtocol)
		}
	})
}
