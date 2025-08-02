package specs

func WithChunkedEncoding(header *Header) *Header {
	header.Set("Transfer-Encoding", "chunked")
	return header
}
