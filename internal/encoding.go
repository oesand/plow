package internal

import "github.com/oesand/giglet/specs"

func IsChunkedEncoding(header *specs.Header) (bool, error) {
	if te, has := header.TryGet("Transfer-Encoding"); has {
		switch te {
		case "chunked":
			return true, nil
		default:
			return false, specs.ErrUnknownTransferEncoding
		}
	}
	return false, nil
}

func IsKnownContentEncoding(e string) bool {
	switch e {
	case specs.ContentEncodingUndefined, specs.ContentEncodingGzip,
		specs.ContentEncodingZstd, specs.ContentEncodingDeflate:
		return true
	}
	return false
}
