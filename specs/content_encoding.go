package specs

type ContentEncoding string

const (
	UnknownContentEncoding ContentEncoding = ""
	GzipContentEncoding    ContentEncoding = "gzip"
)
