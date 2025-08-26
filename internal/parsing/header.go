package parsing

import (
	"bufio"
	"context"
	"github.com/oesand/plow/internal/stream"
	"github.com/oesand/plow/specs"
	"io"
	"strings"
)

func ParseHeaders(ctx context.Context, reader *bufio.Reader, lineLimit int64, totalLimit int64) (*specs.Header, error) {
	var totalLen int64

	// The first line cannot start with a leading space.
	if buf, err := reader.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		line, err := stream.ReadBufferLine(reader, lineLimit)
		if err != nil {
			return nil, err
		}
		totalLen = int64(len(line))
		if totalLimit > 0 && totalLen > lineLimit {
			return nil, specs.ErrTooLarge
		}

		return nil, specs.NewOpError("parsing", "malformed header initial line")
	}

	header := specs.NewHeader()

	var key, value []byte
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		line, err := stream.ReadBufferLine(reader, lineLimit)
		if err != nil && err != io.EOF {
			return nil, err
		} else if len(line) == 0 || err == io.EOF {
			if key != nil {
				applyKVHeader(header, string(key), string(value))
			}
			return header, nil
		}

		if len(line) < 2 {
			continue
		}
		totalLen += int64(len(line))
		if totalLimit > 0 && totalLen > lineLimit {
			return nil, specs.ErrTooLarge
		}

		k, v, ok := parseHeaderKVLine(line)
		if !ok {
			key, value = nil, nil
			continue
		}

		if k == nil {
			if key == nil {
				continue
			}
			value = append(value, ' ')
			value = append(value, v...)
			continue
		}

		if key != nil {
			applyKVHeader(header, string(key), string(value))
		}
		key, value = k, v
	}
}

func applyKVHeader(header *specs.Header, key, value string) {
	if strings.EqualFold(key, "Cookie") {
		for cookieKey, cookieVal := range ParseCookieHeader(value) {
			header.SetCookieValue(cookieKey, cookieVal)
		}
	} else if strings.EqualFold(key, "Set-Cookie") {
		cookie := ParseSetCookieHeader(value)
		if cookie != nil {
			header.SetCookie(*cookie)
		}
	} else {
		header.Set(key, value)
	}
}

func parseHeaderKVLine(line []byte) ([]byte, []byte, bool) {
	var k, v []byte

	var writeVal bool
	capNext := true // capitalize first letter or anything after space

	for i, b := range line {
		if !writeVal {
			if 'a' <= b && b <= 'z' && capNext {
				b -= 32 // to upper
			} else if 'A' <= b && b <= 'Z' && !capNext {
				b += 32 // to lowercase
			}

			switch b {
			case ' ', '\t':
				if i == 0 {
					writeVal = true
					k = nil
				}
				continue
			case ':':
				if len(k) == 0 {
					return nil, nil, false
				}
				writeVal = true
				capNext = false
				continue
			case '-', '_':
				capNext = true
			default:
				capNext = false
			}

			if !isTokenChar(b) {
				return nil, nil, false
			}

			k = append(k, b)
		} else {
			switch b {
			case ' ', '\t':
				continue
			}

			var e int
			for e = len(line) - 1; e > i; e-- {
				switch line[e] {
				case ' ', '\t':
					continue
				}
				break
			}

			v = make([]byte, e-i+1)
			copy(v, line[i:e+1])
			break
		}
	}
	if !writeVal && len(v) == 0 {
		return nil, nil, false
	}
	if v == nil {
		v = make([]byte, 0)
	}
	return k, v, writeVal
}
