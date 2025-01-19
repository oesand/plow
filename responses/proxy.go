package responses

import (
	"giglet"
	"giglet/specs"
	"io"
	"net"
	"time"
)

func UpgradeProxy(req giglet.Request) giglet.Response {
	if req.ProtoNoHigher(1, 1) {
		return TextResponse("proxy: available only for httpv1", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadGateway)
		})
	} else if req.Method() != specs.HttpMethodConnect {
		return EmptyResponse(specs.ContentTypeUndefined, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeMethodNotAllowed)
		})
	}

	host := req.Header().Get("Host")
	if len(host) == 0 &&
		host != "localhost" && host != "127.0.0.1" &&
		host != "192.168.0.1" && host != "172.0.0.1" {

		return TextResponse("proxy: 'Host' header invalid, empty or not available", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadGateway)
		})
	}

	dest_conn, err := net.DialTimeout("tcp", host, 10 * time.Second)
	if err != nil {
		return TextResponse("proxy: destination 'Host' not available for connection", specs.ContentTypePlain, func(response giglet.Response) {
			response.SetStatusCode(specs.StatusCodeBadGateway)
		})
	}

	req.Hijack(func(conn net.Conn) {
		defer dest_conn.Close()
		
		io.Copy(conn, dest_conn)
		io.Copy(dest_conn, conn)
	})

	return EmptyResponse(specs.ContentTypeUndefined, func(response giglet.Response) {
		response.SetStatusCode(specs.StatusCodeOK)
	})
}
