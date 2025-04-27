package parsing

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
	"golang.org/x/net/http/httpguts"
)

var (
	rawColon       = []byte(": ")
	errorCancelled = &specs.GigletError{Err: errors.New("cancelled")}
)

func ParseHeaders(ctx context.Context, reader *bufio.Reader, lineLimit int64, totalLimit int64) (map[string]string, map[string]*specs.Cookie, error) {
	select {
	case <-ctx.Done():
		return nil, nil, errorCancelled
	default:
	}

	var totalLen int64

	// The first line cannot start with a leading space.
	if buf, err := reader.Peek(1); err == nil && (buf[0] == ' ' || buf[0] == '\t') {
		line, err := utils.ReadBufferLine(reader, lineLimit)
		if err != nil {
			return nil, nil, err
		}
		totalLen = int64(len(line))
		if totalLimit > 0 && totalLen > lineLimit {
			return nil, nil, &specs.GigletError{
				Op:  "headers/server",
				Err: fmt.Errorf("too large (%d > %d)", totalLimit, lineLimit),
			}
		}

		return nil, nil, &specs.GigletError{
			Op:  "headers/server",
			Err: fmt.Errorf("malformed header initial line: %s", line),
		}
	}

	headers := map[string]string{}
	cookies := map[string]*specs.Cookie{}

	var key, value []byte
	for {
		select {
		case <-ctx.Done():
			return nil, nil, errorCancelled
		default:
		}

		line, err := utils.ReadBufferLine(reader, lineLimit)
		if err != nil {
			return nil, nil, &specs.GigletError{
				Op:  "headers/server",
				Err: err,
			}
		} else if len(line) == 0 {
			return headers, cookies, nil
		}

		line = bytes.TrimLeft(line, " ")
		if len(line) < 2 {
			continue
		}
		totalLen += int64(len(line))
		if totalLimit > 0 && totalLen > lineLimit {
			return nil, nil, &specs.GigletError{
				Op:  "headers/server",
				Err: fmt.Errorf("too large (%d > %d)", totalLimit, lineLimit),
			}
		}

		if value != nil && len(value) != 0 && line[0] == '\t' {
			value = append(value, line[1:]...)
			continue
		}

		var ok bool
		key, value, ok = bytes.Cut(line, rawColon)
		if !ok || len(key) == 0 || len(value) == 0 {
			continue
		}

		headerKey, headerVal := utils.BufferToString(utils.TitleCaseBytes(key)), string(value)
		if httpguts.ValidHeaderFieldName(headerKey) && httpguts.ValidHeaderFieldValue(headerVal) {
			if headerKey == "Cookie" {
				for cookieKey, cookieVal := range ParseCookieHeader(headerVal) {
					cookies[cookieKey] = &specs.Cookie{
						Name:  cookieKey,
						Value: cookieVal,
					}
				}
			} else if headerKey == "Set-Cookie" {
				cookie := ParseSetCookieHeader(headerVal)
				if cookie != nil {
					cookies[cookie.Name] = cookie
				}
			} else {
				headers[headerKey] = headerVal
			}
		}
		key, value = nil, nil
	}
}
