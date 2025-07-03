package specs

import (
	"encoding/base64"
	"github.com/oesand/giglet/internal"
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

func BasicAuthHeader(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString(internal.StringToBuffer(auth))
}
