package proxy

import (
	"github.com/oesand/plow/specs"
	"net"
	"strconv"
)

const (
	DefaultHttpPort   uint16 = 8080
	DefaultHttpsPort  uint16 = 443
	DefaultSocks5Port uint16 = 1080
)

var SchemeDefaultPortMap = map[string]uint16{
	"http":    DefaultHttpPort,
	"https":   DefaultHttpsPort,
	"socks5":  DefaultSocks5Port,
	"socks5h": DefaultSocks5Port,
}

type Creds struct {
	Username string
	Password string
}

func WithAuthHeader(header *specs.Header, username, password string) *specs.Header {
	header.Set("Proxy-Authorization", specs.BasicAuthHeader(username, password))
	return header
}

type ResolvedAddr struct {
	Net    string
	Domain string
	IP     net.IP
	Port   int
}

func (a *ResolvedAddr) Network() string { return a.Net }

func (a *ResolvedAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	port := strconv.Itoa(a.Port)
	if a.IP == nil {
		return a.Domain + ":" + port
	}
	return a.IP.String() + ":" + port
}
