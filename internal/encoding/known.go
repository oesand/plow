package encoding

import "github.com/oesand/giglet/specs"

const DefaultAcceptEncoding = specs.ContentEncodingGzip + ", " +
	specs.ContentEncodingDeflate + ", " + specs.ContentEncodingBrotli

func IsChunkedTransfer(header *specs.Header) (bool, error) {
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

func IsKnownEncoding(contentEncoding string) bool {
	switch contentEncoding {
	case specs.ContentEncodingGzip, specs.ContentEncodingDeflate, specs.ContentEncodingBrotli:
		return true
	}
	return false
}
