package responses

import (
	"giglet"
	"giglet/specs"
	"net"
	"strings"
)

func UpgradeWebSocket(req giglet.Request, conf *WebSocketConf, handler WebSocketHandler) giglet.Response {
	if req.Method() != specs.HttpMethodGet {
		return TextResponse("websocket: upgrading required request method - GET", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeMethodNotAllowed)
		})
	} else if !strings.EqualFold(req.Header().Get("Connection"), "Upgrade") {
		return TextResponse("websocket: 'Upgrade' token not found in 'Connection' header", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadRequest)
		})
	} else if !strings.EqualFold(req.Header().Get("Upgrade"), "websocket") {
		return TextResponse("websocket: 'websocket' token not found in 'Upgrade' header", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadRequest)
		})
	} else if req.Header().Get("Sec-Websocket-Version") != "13" {
		return TextResponse("websocket: supports only websocket 13 version", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeNotImplemented)
		})
	}
	
	challengeKey := req.Header().Get("Sec-Websocket-Key")
	if len(challengeKey) == 0 {
		return TextResponse("websocket: not a websocket handshake: `Sec-WebSocket-Key' header is missing or blank", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadRequest)
		})
	}
	if conf == nil {
		conf = &WebSocketConf{}
	}
	req.Hijack(func(conn net.Conn) {
		reader := bufioReaderPool.Get(conn)
		defer bufioReaderPool.Put(reader)

		handler(&WebSocketConn{
			request: req,
			conn: conn,
			reader: *reader,
			conf: *conf,
		})
	})

	return EmptyResponse(specs.ContentTypeUndefined, func(resp giglet.Response) {
		resp.SetStatusCode(specs.StatusCodeSwitchingProtocols)
		resp.Header().Set("Upgrade", "websocket")
		resp.Header().Set("Connection", "Upgrade")
		resp.Header().Set("Sec-WebSocket-Accept", specs.ComputeWebSocketAcceptKey(challengeKey))
		if conf.EnableCompression {
			if ext := req.Header().Get("Sec-WebSocket-Extensions"); 
				len(ext) > 0 && strings.Contains(ext, "permessage-deflate") {

				resp.Header().Set("Sec-WebSocket-Extensions", "permessage-deflate; server_no_context_takeover; client_no_context_takeover");
			}
		}
	})
}
