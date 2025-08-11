package specs

import (
	"encoding/base64"
	"errors"
	"github.com/oesand/giglet/internal"
	"strings"
)

// TimeFormat is the format used for HTTP date headers.
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// BasicAuthHeader creates a Basic Authentication header from a username and password.
func BasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString(internal.StringToBuffer(auth))
}

// BearerAuthHeader creates a Bearer Authentication header from a token.
func BearerAuthHeader(token string) string {
	return "Bearer " + token
}

// WithBearerAuthHeader adds a Bearer Authentication header to the provided [Header].
func WithBearerAuthHeader(header *Header, token string) *Header {
	header.Set("Authorization", BearerAuthHeader(token))
	return header
}

// ParseBasicAuthHeader parses a Basic Authentication header and returns the username and password.
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

// ParseBearerAuthHeader parses a Bearer Authentication header and returns the token.
func ParseBearerAuthHeader(header string) (token string, err error) {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("missing Bearer prefix")
	}

	return strings.TrimPrefix(header, prefix), nil
}
