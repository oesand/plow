package client

import (
	"golang.org/x/net/idna"
	"strconv"
	"strings"
	"unicode"
)

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

// TODO removeZone removes IPv6 zone identifier from host.
// E.g., "[fe80::1%en0]:8080" to "[fe80::1]:8080"
func removeZone(host string) string {
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
