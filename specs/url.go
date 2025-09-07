package specs

import (
	"errors"
	"github.com/oesand/plow/internal/plain"
	"strconv"
	"strings"
)

// MustParseUrl is a helper function that parses a URL string and panics if it fails.
func MustParseUrl(url string) *Url {
	ur, err := ParseUrl(url)
	if err != nil {
		panic(err)
	}
	return ur
}

// ParseUrl parses a URL string and returns a Url object.
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
			return nil, errors.New("url: invalid control character in url")
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
				return nil, errors.New("url: invalid scheme suffix")
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
			return nil, errors.New("url: username must not be empty when passed")
		}
	} else {
		url = strings.TrimSuffix(url, "@")
	}

	// Parse host:port
	portIndex := strings.LastIndex(url, ":")
	if strings.HasPrefix(url, "[") {
		i := strings.Index(url, "]")
		if i < 0 {
			return nil, errors.New("url: missing ']' in host")
		}
		if i > portIndex {
			portIndex = -1
		}
	}

	if portIndex == 0 {
		return nil, errors.New("url: empty host when port passed")
	} else if portIndex > 0 {
		if portIndex == len(url)-1 {
			return nil, errors.New("url: empty port")
		}

		host, port := url[:portIndex], url[portIndex+1:]
		obj.Host = host

		portNum, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, errors.New("url: cannot parse port")
		}
		obj.Port = uint16(portNum)
	} else {
		obj.Host = url
	}

	if obj.Host == "" {
		if obj.Scheme != "" {
			return nil, errors.New("url: host required when scheme passed")
		}
		if obj.Username != "" {
			return nil, errors.New("url: host required when username passed")
		}
	}

	// Escaping
	if len(obj.Host) > 0 {
		host := obj.Host
		if strings.HasSuffix(host, "]") && !strings.HasPrefix(obj.Host, "[") {
			return nil, errors.New("url: missing '[' in host")
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

// Url represents URL with its components.
type Url struct {
	// Url raw component before escaping
	Scheme, Username, Password,
	Host, Path, Fragment string
	Port uint16

	// Query map of raw unescaped values represents of url Query.
	Query Query

	// PathSegments slice of raw unescaped components of Path.
	//
	// ParseUrl will ignore this field.
	// if PathSegments provided String will use this and escape every segment.
	PathSegments []string
}

// EscapedPath returns the escaped form of Path.
// In general there are multiple possible escaped forms of any path.
// EscapedPath returns PathSegments when it is provided.
// Otherwise, EscapedPath ignores PathSegments and computes an escaped
// form of Path.
func (url *Url) EscapedPath() string {
	if segments := url.PathSegments; segments != nil {
		if len(segments) > 0 {
			var builder strings.Builder
			for _, segment := range segments {
				builder.WriteByte('/')
				builder.WriteString(plain.EscapeUrl(segment, plain.EscapingPathSegment))
			}
			return builder.String()
		}
	} else if path := url.Path; path != "" {
		escaped := plain.EscapeUrl(path, plain.EscapingPath)
		if path[0] != '/' {
			escaped = "/" + escaped
		}
		return escaped
	}
	return ""
}

// String returns the string representation of the URL.
// It constructs the URL from its components, escaping them as necessary.
func (url *Url) String() string {
	var builder strings.Builder

	if url.Host != "" {
		if url.Scheme != "" {
			builder.WriteString(url.Scheme)
			builder.WriteString("://")
		}
		if url.Username != "" {
			builder.WriteString(plain.EscapeUrl(url.Username, plain.EscapingUserPassword))
			if url.Password != "" {
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

	if path := url.EscapedPath(); path != "" {
		builder.WriteString(path)
	} else if url.Host != "" {
		builder.WriteByte('/')
	}

	if url.Query.Any() {
		builder.WriteByte('?')
		builder.WriteString(url.Query.String())
	}

	if url.Fragment != "" {
		builder.WriteByte('#')
		builder.WriteString(plain.EscapeUrl(url.Fragment, plain.EscapingFragment))
	}

	return builder.String()
}
