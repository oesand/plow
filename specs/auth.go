package specs

import (
	"encoding/base64"
	"errors"
	"github.com/oesand/giglet/internal"
	"strings"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func BasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(internal.StringToBuffer(auth))
}

func BearerAuthHeader(token string) string {
	return "Bearer " + token
}

func WithBearerAuthHeader(header *Header, token string) *Header {
	header.Set("Authorization", BearerAuthHeader(token))
	return header
}

func ParseBasicAuthHeader(header string) (username, password string, err error) {
	const prefix = "Basic "

	if !strings.HasPrefix(header, prefix) {
		return "", "", errors.New("missing Basic prefix")
	}

	// Decode the base64 part
	encoded := strings.TrimPrefix(header, prefix)
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", errors.New("cannot decode base64")
	}

	// Split into username and password
	parts := strings.SplitN(string(decodedBytes), ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid format, expected username:password")
	}

	return parts[0], parts[1], nil
}

func ParseBearerAuthHeader(header string) (token string, err error) {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("missing Bearer prefix")
	}

	return strings.TrimPrefix(header, prefix), nil
}
