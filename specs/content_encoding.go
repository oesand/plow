package specs

// ContentEncoding represents the content encoding types used in HTTP responses.
//
// These constants are used to specify the encoding of the response body,
// allowing clients to understand how to decode the content.
// The values are based on the IANA HTTP Content-Encoding registry.
//
// Reference: https://www.iana.org/assignments/http-parameters/http-parameters.xhtml
const (
	ContentEncodingGzip    = "gzip"
	ContentEncodingDeflate = "deflate"
	ContentEncodingBrotli  = "br"
)
