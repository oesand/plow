package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	socksVersion5   byte = 0x05
	socksCmdConnect byte = 0x01

	socksNoAuthFlag      byte = 0x00
	socksAuthByCredsFlag byte = 0x02

	socksAuthCredsVersion byte = 0x01
	socksAuthSucceeded    byte = 0x00

	socksAddrTypeIPv4 byte = 0x01
	socksAddrTypeFQDN byte = 0x03
	socksAddrTypeIPv6 byte = 0x04
)

func DialSocks5(conn net.Conn, host string, port uint16, creds *Creds) (net.Addr, error) {
	if creds != nil && (len(creds.Username) == 0 || len(creds.Username) > 255 || len(creds.Password) > 255) {
		return nil, errors.New("socks5: invalid username/password")
	}

	buf := make([]byte, 0, 6+len(host)) // the size here is just an estimate
	buf = append(buf, socksVersion5)

	if creds == nil {
		buf = append(buf, 1, socksNoAuthFlag)
	} else {
		buf = append(buf, byte(2))
		buf = append(buf, socksNoAuthFlag)
		buf = append(buf, socksAuthByCredsFlag)
	}

	var err error
	if _, err = conn.Write(buf); err != nil {
		return nil, err
	}

	if _, err = io.ReadFull(conn, buf[:2]); err != nil {
		return nil, err
	}

	if ver := buf[0]; ver != socksVersion5 {
		return nil, errors.New(fmt.Sprintf("socks5: unexpected protocol version: %d", int(ver)))
	}

	authMethod := buf[1]
	switch authMethod {
	case socksNoAuthFlag:
		break
	case socksAuthByCredsFlag:
		if creds == nil {
			return nil, errors.New("socks5: authentication required")
		}
		credsBuf := []byte{socksAuthCredsVersion}
		credsBuf = append(credsBuf, byte(len(creds.Username)))
		credsBuf = append(credsBuf, creds.Username...)
		credsBuf = append(credsBuf, byte(len(creds.Password)))
		credsBuf = append(credsBuf, creds.Password...)

		if _, err = conn.Write(credsBuf); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(conn, credsBuf[:2]); err != nil {
			return nil, err
		}
		if credsBuf[0] != socksAuthCredsVersion {
			return nil, errors.New("socks5: invalid username/password version")
		}
		if credsBuf[1] != socksAuthSucceeded {
			return nil, errors.New("socks5: username/password authentication failed")
		}
	default:
		return nil, errors.New("socks5: no acceptable authentication methods")
	}

	buf = buf[:0]
	buf = append(buf, socksVersion5, socksCmdConnect, 0)
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			buf = append(buf, socksAddrTypeIPv4)
			buf = append(buf, ip4...)
		} else if ip6 := ip.To16(); ip6 != nil {
			buf = append(buf, socksAddrTypeIPv6)
			buf = append(buf, ip6...)
		} else {
			return nil, errors.New("socks5: unknown address type")
		}
	} else {
		if len(host) > 255 {
			return nil, errors.New("socks5: FQDN too long")
		}
		buf = append(buf, socksAddrTypeFQDN)
		buf = append(buf, byte(len(host)))
		buf = append(buf, host...)
	}
	buf = append(buf, byte(port>>8), byte(port))

	if _, err = conn.Write(buf); err != nil {
		return nil, err
	}

	if _, err = io.ReadFull(conn, buf[:4]); err != nil {
		return nil, err
	}
	if buf[0] != socksVersion5 {
		return nil, errors.New(fmt.Sprintf("socks5: unexpected protocol version %d", int(buf[0])))
	}
	if replyCode := buf[1]; replyCode != socksAuthSucceeded {
		return nil, errors.New(fmt.Sprintf("socks5: reply error: %s", socksReplyCodeToError(replyCode)))
	}
	if buf[2] != 0 {
		return nil, errors.New("socks5: non-zero reserved field")
	}

	l := 2
	var addr = ResolvedAddr{Net: "socks"}
	switch buf[3] {
	case socksAddrTypeIPv4:
		l += net.IPv4len
		addr.IP = make(net.IP, net.IPv4len)
	case socksAddrTypeIPv6:
		l += net.IPv6len
		addr.IP = make(net.IP, net.IPv6len)
	case socksAddrTypeFQDN:
		if _, err = io.ReadFull(conn, buf[:1]); err != nil {
			return nil, err
		}
		l += int(buf[0])
	default:
		return nil, errors.New(fmt.Sprintf("socks5: unknown address type %d", int(buf[3])))
	}

	if cap(buf) < l {
		buf = make([]byte, l)
	} else {
		buf = buf[:l]
	}
	if _, err = io.ReadFull(conn, buf); err != nil {
		return nil, err
	}

	if addr.IP != nil {
		copy(addr.IP, buf)
	} else {
		addr.Domain = string(buf[:len(buf)-2])
	}
	addr.Port = int(buf[len(buf)-2])<<8 | int(buf[len(buf)-1])

	return &addr, nil
}

func socksReplyCodeToError(code byte) string {
	switch code {
	case socksAuthSucceeded:
		return "succeeded"
	case 0x01:
		return "general SOCKS server failure"
	case 0x02:
		return "connection not allowed by ruleset"
	case 0x03:
		return "network unreachable"
	case 0x04:
		return "host unreachable"
	case 0x05:
		return "connection refused"
	case 0x06:
		return "TTL expired"
	case 0x07:
		return "command not supported"
	case 0x08:
		return "address type not supported"
	default:
		return fmt.Sprintf("unknown code: %d", int(code))
	}
}
