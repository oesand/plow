package server_ops

func TlsRecordHeaderLikeHTTP(hdr [5]byte) bool {
	switch string(hdr[:]) {
	case "GET /", "HEAD ", "POST ", "PUT /", "OPTIO", "DELET", "CONNE", "PATCH", "TRACE":
		return true
	}
	return false
}
