package encoding

import "github.com/oesand/giglet/specs"

func IsChunkedEncoding(header *specs.Header) (bool, error) {
	if te, has := header.TryGet("Transfer-Encoding"); has {
		switch te {
		case "chunked":
			if header.Has("Content-Length") {
				header.Del("Content-Length")
			}
			return true, nil
		default:
			return false, specs.ErrUnknownTransferEncoding
		}
	}
	return false, nil
}

func IsKnownEncoding(e string) bool {
	switch e {
	case specs.UnknownContentEncoding, specs.GzipContentEncoding:
		return true
	}
	return false
}
