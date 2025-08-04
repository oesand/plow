package proxy

import (
	"context"
	"errors"
	"github.com/oesand/giglet/internal/client"
	"github.com/oesand/giglet/internal/stream"
	"github.com/oesand/giglet/specs"
	"net"
	"net/http"
)

func DialHttps(conn net.Conn, host string, port uint16, creds *Creds) error {
	if creds != nil && len(creds.Username) == 0 {
		return errors.New("https: invalid username")
	}

	address := client.HostHeader(host, port, true)

	header := specs.NewHeader()
	header.Set("Host", address)
	if creds != nil {
		WithAuthHeader(header, creds.Username, creds.Password)
	}

	_, err := client.WriteRequestHead(conn, specs.HttpMethodConnect, address, nil, header)
	if err != nil {
		return err
	}

	reader := stream.DefaultBufioReaderPool.Get(conn)
	defer stream.DefaultBufioReaderPool.Put(reader)

	resp, err := client.ReadResponse(context.Background(), reader, 1024, 4*1024)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New("https: invalid status code")
	}

	return nil
}
