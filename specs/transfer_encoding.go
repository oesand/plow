package specs

// WithChunkedEncoding sets the Transfer-Encoding header to "chunked".
func WithChunkedEncoding(header *Header) *Header {
	header.Set("Transfer-Encoding", "chunked")
	return header
}
