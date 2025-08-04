package client

import (
	"golang.org/x/net/idna"
	"strconv"
	"strings"
	"unicode"
)

// HostPort concat host and port.
// host must be idna formatted
func HostPort(host string, port uint16) string {
	return host + ":" + strconv.FormatUint(uint64(port), 10)
}

func IdnaHost(host string) string {
	if !isAscii(host) {
		if v, err := idna.Lookup.ToASCII(host); err == nil {
			host = v
		}
	}
	return host
}

func isAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// HostHeader compute valid host header.
// host must be idna formatted
func HostHeader(host string, port uint16, isProxy bool) string {
	host = removeIPv6Zone(host)
	if !isProxy && (port == 80 || port == 443) {
		return host
	}
	return HostPort(host, port)
}

func removeIPv6Zone(host string) string {
	if !strings.HasPrefix(host, "[") {
		return host
	}
	i := strings.LastIndex(host, "]")
	if i < 0 {
		return host
	}
	j := strings.LastIndex(host[:i], "%")
	if j < 0 {
		return host
	}
	return host[:j] + host[i:]
}
