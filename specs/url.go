package specs

import (
	"errors"
	"github.com/oesand/giglet/internal/plain"
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
	switch url {
	case "":
		return &Url{}, nil
	case "/":
		return &Url{Path: url}, nil
	}

	// invalid control character
	for i := 0; i < len(url); i++ {
		c := url[i]
		if c < ' ' || c == 0x7f {
			return nil, errors.New("invalid control character in url")
		}
	}

	obj := &Url{}

	// Parse scheme
	end := len(url) - 1
	for i, c := range url {
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' ||
			(i > 0 && ('0' <= c && c <= '9' || c == '+' || c == '-' || c == '.')) {
			continue
		}

		if i+2 <= end && url[i] == ':' && url[i+1] == '/' {
			if 0 == i || i+2 == end || url[i+2] != '/' {
				return nil, errors.New("invalid scheme suffix")
			}

			obj.Scheme = url[:i]
			url = url[i+3:]
		}

		break
	}

	// Parse fragment
	if rest, fragment, ok := strings.Cut(url, "#"); ok {
		url = rest
		obj.Fragment = fragment
	} else {
		url = strings.TrimSuffix(url, "#")
	}

	// Parse query
	if rest, query, ok := strings.Cut(url, "?"); ok {
		url = rest
		obj.Query = ParseQuery(query)
	} else {
		url = strings.TrimSuffix(url, "?")
	}

	// Parse path
	if rest, path, ok := strings.Cut(url, "/"); ok {
		url = rest
		if path == "" {
			obj.Path = "/"
		} else {
			obj.Path = "/" + path
		}
	} else {
		url = strings.TrimSuffix(url, "/")
	}

	// Parse username & password
	if raw, rest, ok := strings.Cut(url, "@"); ok {
		url = rest
		if username, password, ok := strings.Cut(raw, ":"); ok {
			obj.Username = username
			obj.Password = password
		} else {
			obj.Username = raw
		}

		if obj.Username == "" {
			return nil, errors.New("username must not be empty when passed")
		}
	} else {
		url = strings.TrimSuffix(url, "@")
	}

	// Parse host:port
	portIndex := strings.LastIndex(url, ":")
	if strings.HasPrefix(url, "[") {
		i := strings.Index(url, "]")
		if i < 0 {
			return nil, errors.New("missing ']' in host")
		}
		if i > portIndex {
			portIndex = -1
		}
	}

	if portIndex == 0 {
		return nil, errors.New("empty host when port passed")
	} else if portIndex > 0 {
		if portIndex == len(url)-1 {
			return nil, errors.New("empty port")
		}

		host, port := url[:portIndex], url[portIndex+1:]
		obj.Host = host

		portNum, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, errors.New("cannot parse port")
		}
		obj.Port = uint16(portNum)
	} else {
		obj.Host = url
	}

	if obj.Host == "" {
		if obj.Scheme != "" {
			return nil, errors.New("host required when scheme passed")
		}
		if obj.Username != "" {
			return nil, errors.New("host required when username passed")
		}
	}

	// Escaping
	if len(obj.Host) > 0 {
		host := obj.Host
		if strings.HasSuffix(host, "]") && !strings.HasPrefix(obj.Host, "[") {
			return nil, errors.New("missing '[' in host")
		}

		if strings.HasPrefix(host, "[") {
			// RFC 6874 defines that %25 (%-encoded percent) introduces
			// the zone identifier, and the zone identifier can use basically
			// any %-encoding it likes. That's different from the host, which
			// can only %-encode non-ASCII bytes.
			// We do impose some restrictions on the zone, to avoid stupidity
			// like newlines.
			zoneIndex := strings.Index(host, "%25")
			if zoneIndex >= 0 {
				host1, err := plain.UnEscapeUrl(host[:zoneIndex], plain.EscapingHost)
				if err != nil {
					return nil, err
				}
				host2, err := plain.UnEscapeUrl(host[zoneIndex:len(host)-1], plain.EscapingZone)
				if err != nil {
					return nil, err
				}
				obj.Host = host1 + host2
			} else if unesc, err := plain.UnEscapeUrl(host, plain.EscapingHost); err == nil {
				obj.Host = unesc
			} else {
				return nil, err
			}
		} else if unesc, err := plain.UnEscapeUrl(host, plain.EscapingHost); err == nil {
			obj.Host = unesc
		} else {
			return nil, err
		}
	}

	if unesc, err := plain.UnEscapeUrl(obj.Username, plain.EscapingUserPassword); err == nil {
		obj.Username = unesc
	} else {
		return nil, err
	}

	if unesc, err := plain.UnEscapeUrl(obj.Password, plain.EscapingUserPassword); err == nil {
		obj.Password = unesc
	} else {
		return nil, err
	}

	if unesc, err := plain.UnEscapeUrl(obj.Path, plain.EscapingPath); err == nil {
		obj.Path = unesc
	} else {
		return nil, err
	}

	if unesc, err := plain.UnEscapeUrl(obj.Fragment, plain.EscapingFragment); err == nil {
		obj.Fragment = unesc
	} else {
		return nil, err
	}

	return obj, nil
}

type Url struct {
	Scheme, Username, Password,
	Host, Path, Fragment string
	Port uint16

	Query Query
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

		builder.WriteString(plain.EscapeUrl(url.Host, plain.EscapingHost))

		if url.Port > 0 {
			builder.WriteByte(':')
			builder.Write(strconv.AppendUint(nil, uint64(url.Port), 10))
		}
	}

	if len(url.Path) > 0 {
		builder.WriteString(plain.EscapeUrl(url.Path, plain.EscapingPath))
	}

	if url.Query.Any() {
		builder.WriteByte('?')
		builder.WriteString(url.Query.String())
	}

	if len(url.Fragment) > 0 {
		builder.WriteByte('#')
		builder.WriteString(plain.EscapeUrl(url.Fragment, plain.EscapingFragment))
	}

	return builder.String()
}
