package proxy

import (
	"context"
	"errors"
	"github.com/oesand/plow/internal/client_ops"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"net"
	"net/http"
)

func DialHttps(conn net.Conn, host string, port uint16, creds *Creds) error {
	if creds != nil && len(creds.Username) == 0 {
		return errors.New("https: invalid username")
	}

	address := client_ops.HostHeader(host, port, true)

	header := specs.NewHeader()
	header.Set("Host", address)
	if creds != nil {
		WithAuthHeader(header, creds.Username, creds.Password)
	}

	_, err := client_ops.WriteRequestHead(conn, specs.HttpMethodConnect, address, nil, header)
	if err != nil {
		return err
	}

	reader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(reader)

	resp, err := client_ops.ReadResponse(context.Background(), reader, 1024, 4*1024)
	if err != nil {
		return err
	}

	if code := resp.StatusCode(); code != http.StatusOK {
		return errors.New("https: invalid status code: " + string(code.Formatted()))
	}

	return nil
}
