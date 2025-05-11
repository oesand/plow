package specs

import (
	"github.com/oesand/giglet/internal/utils/plain"
	"strconv"
	"strings"
)

func MustParseUrl(url string) *Url {
	ur, err := ParseUrl(url)
	if err != nil {
		panic(err)
	}
	return ur
}

func MustParseUrlQuery(url string, query Query) *Url {
	obj, err := ParseUrl(url)
	if err != nil {
		panic(err)
	}
	if obj.Query == nil || len(obj.Query) == 0 {
		obj.Query = query
	}
	return obj
}

func ParseUrl(url string) (*Url, error) {
	obj := &Url{}
	if len(url) == 0 {
		return obj, nil
	}

	if url == "/" {
		obj.Path = url
		return obj, nil
	}

	if rest, raw, ok := strings.Cut(url, "#"); ok {
		url = rest
		if hash, err := plain.UnEscapeUrl(raw, plain.EscapingFragment); err == nil {
			obj.Hash = hash
		} else {
			obj.Hash = raw
		}
	}

	if url == "" {
		return obj, nil
	}

	// Scheme[0], Username[1], Password[2], Host[3], Port[4], Path[5], Query[6], Hash[7]
	const (
		stepScheme = iota
		stepHost
		stepPort
		stepPath
		stepQuery
	)

	var mark, step int
	end := len(url) - 1

	if url[0] == '/' {
		step = stepPath
	}

	for i := 0; i < len(url); i++ {
		c := url[i]
		switch {
		// invalid control character
		case c < ' ' || c == 0x7f:
			return nil, ErrInvalidFormat

		// read 'scheme'
		case step == stepScheme && i+2 <= end && url[i] == ':' && url[i+1] == '/':
			if 0 == i || i+2 == end || url[i+2] != '/' {
				return nil, ErrInvalidFormat
			}

			step = stepHost // goto 'host'
			obj.Scheme = url[mark:i]
			i += 2
			mark = i + 1
			continue

		// read 'host'
		case step == stepScheme || step == stepHost:
			switch {
			case i == end:
				if len(url)-mark < 1 {
					return nil, ErrInvalidFormat
				}
				obj.Host = url[mark:]
				step = stepHost // exit with ends on 'host'

			default:
				// invalid 'scheme' characters - force 'host'
				if step == stepScheme &&
					!('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' ||
						(i > 0 && ('0' <= c && c <= '9' || c == '+' || c == '-' || c == '.'))) {

					step = stepHost // goto 'host'
				}

				switch c {
				case ':':
					if url[mark] == '[' && url[i-1] != ']' {
						continue
					}
					step = stepPort // goto 'port'
				case '/':
					step = stepPath // goto 'path'
				case '?':
					step = stepQuery // goto 'query'
				case '@': // read as 'user'
					if i-mark <= 1 || obj.Username != "" {
						return nil, ErrInvalidFormat
					}
					obj.Username = url[mark:i]
					step = stepHost // goto 'host'
					mark = i + 1
					continue
				default:
					continue
				}

				if i-mark < 1 {
					return nil, ErrInvalidFormat
				}

				obj.Host = url[mark:i]
				mark = i + 1
			}

		// read 'port'
		case step == stepPort:
			switch {
			case i == end:
				if len(url)-mark < 1 {
					return nil, ErrInvalidFormat
				}
				if !obj.setPort(url[mark:]) {
					return nil, ErrInvalidFormat
				}
			default:
				switch c {
				case '/':
					step = stepPath // goto 'path'
				case '?':
					step = stepQuery // goto 'query'
				case '@': // read as 'user & pass'
					if i-mark <= 1 || obj.Username != "" || obj.Password != "" {
						return nil, ErrInvalidFormat
					}
					obj.Username = obj.Host
					obj.Password = url[mark:i]
					obj.Host = ""
					step = stepHost // goto 'host'
					mark = i + 1
					continue
				default:
					continue
				}
				if i-mark < 1 {
					return nil, ErrInvalidFormat
				}
				if !obj.setPort(url[mark:i]) {
					return nil, ErrInvalidFormat
				}
				mark = i + 1
			}

		// read 'path'
		case step == stepPath:
			switch {
			case i == end || c == '?':
				if mark != 0 {
					mark--
				}

				var text string
				if i == end {
					text = url[mark:]
				} else {
					text = url[mark:i]
				}

				if text == "/" {
					obj.Path = text
				} else if path, err := plain.UnEscapeUrl(text, plain.EscapingPath); err == nil {
					obj.Path = path
				} else {
					obj.Path = text
				}

				mark = i + 1
				if c == '?' {
					step = stepQuery // goto 'query'
				}
			}

		// read 'query'
		case step == stepQuery:
			switch {
			case i == end:
				obj.Query = ParseQuery(url[mark:])
			}
		}
	}

	return obj, nil
}

type Url struct {
	Scheme, Username, Password,
	Host, Path, Hash string
	Port uint16

	Query Query
}

func (url *Url) setPort(val string) bool {
	num, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		return false
	}
	url.Port = uint16(num)
	return true
}

func (url *Url) String() string {
	var builder strings.Builder

	if len(url.Host) > 0 {
		if len(url.Scheme) > 0 {
			builder.WriteString(url.Scheme)
			builder.WriteString("://")
		}
		if len(url.Username) > 0 {
			builder.WriteString(plain.EscapeUrl(url.Username, plain.EscapingUserPassword))
			if len(url.Password) > 0 {
				builder.WriteByte(':')
				builder.WriteString(plain.EscapeUrl(url.Password, plain.EscapingUserPassword))
			}
			builder.WriteByte('@')
		}

		builder.WriteString(url.Host)

		if url.Port > 0 {
			builder.WriteByte(':')
			builder.Write(strconv.AppendUint(nil, uint64(url.Port), 10))
		}
	}

	if len(url.Path) > 0 {
		builder.WriteString(plain.EscapeUrl(url.Path, plain.EscapingPath))
	}

	if url.Query != nil || len(url.Query) > 0 {
		builder.WriteByte('?')
		builder.WriteString(url.Query.String())
	}

	if len(url.Hash) > 0 {
		builder.WriteByte('#')
		builder.WriteString(plain.EscapeUrl(url.Hash, plain.EscapingFragment))
	}

	return builder.String()
}
