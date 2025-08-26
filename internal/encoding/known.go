package encoding

import "github.com/oesand/plow/specs"

const DefaultAcceptEncoding = specs.ContentEncodingGzip + ", " +
	specs.ContentEncodingDeflate + ", " + specs.ContentEncodingBrotli

func IsKnownEncoding(contentEncoding string) bool {
	switch contentEncoding {
	case specs.ContentEncodingGzip, specs.ContentEncodingDeflate, specs.ContentEncodingBrotli:
		return true
	}
	return false
}
