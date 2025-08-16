package parsing

import (
	"github.com/oesand/plow/internal"
	"github.com/oesand/plow/specs"
	"slices"
	"strconv"
)

// parse headline: GET /index.html HTTP/1.0
func ParseClientRequestHeadline(line []byte) (method specs.HttpMethod, url string, major, minor uint16, ok bool) {
	var proto []byte
	for i, b := range line {
		if b == ' ' {
			if method == "" {
				method = specs.HttpMethod(internal.BufferToString(line[:i]))
				if i < 3 {
					break
				}
			} else {
				if i-len(method) <= 1 ||
					len(line)-i != 9 {
					break
				}

				url = internal.BufferToString(line[len(method)+1 : i])
				proto = line[i+1:]
			}
		}
	}
	if method == "" || url == "" || proto == nil {
		return
	}
	major, minor, ok = parseHTTPVersion(proto)
	return
}

// parse headline: HTTP/1.0 200 OK
func ParseServerResponseHeadline(line []byte) (status specs.StatusCode, major, minor uint16, res bool) {
	if len(line) < 14 || !slices.Equal(line[:5], httpVersionPrefix) ||
		line[4] != '/' || line[6] != '.' || line[8] != ' ' || line[12] != ' ' {
		return
	}
	major, minor, res = parseHTTPVersion(line[:8])
	if !res {
		return
	}
	res = false
	code, err := strconv.ParseUint(internal.BufferToString(line[9:12]), 10, 16)
	if err != nil {
		return
	}
	status = specs.StatusCode(code)
	res = true
	return
}
