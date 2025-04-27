package specs

import (
	"github.com/oesand/giglet/internal/utils"
	"iter"
	"mime"
	"strconv"
)

func NewReadOnlyHeader(headers map[string]string, cookies map[string]*Cookie) *ReadOnlyHeader {
	header := &ReadOnlyHeader{headers: headers}
	if media, has := headers["Content-Type"]; has {
		contentType, mediaParams, err := mime.ParseMediaType(media)
		if err != nil {
			header.contentType = ContentType(media)
		} else {
			header.contentType = ContentType(contentType)
			header.mediaParams = mediaParams
		}
		delete(headers, "Content-Type")
	}
	if len, has := headers["Content-Length"]; has {
		length, err := strconv.ParseInt(len, 10, 64)
		if err != nil {
			header.contentLength = -1
		} else {
			header.contentLength = length
		}
		delete(headers, "Content-Length")
	}
	header.cookies = cookies
	return header
}

type ReadOnlyHeader struct {
	_ utils.NoCopy

	contentType   ContentType
	contentLength int64
	mediaParams   map[string]string

	headers map[string]string
	cookies map[string]*Cookie
}

func (header *ReadOnlyHeader) Get(name string) string {
	if header.headers == nil {
		return ""
	} else if name == "Content-Type" {
		return string(header.contentType)
	}

	return header.headers[name]
}

func (header *ReadOnlyHeader) TryGet(name string) (string, bool) {
	if header.headers == nil {
		return "", false
	} else if name == "Content-Type" {
		return string(header.contentType), true
	}

	value, has := header.headers[name]
	return value, has
}

func (header *ReadOnlyHeader) All() iter.Seq2[string, string] {
	if header.headers == nil {
		return utils.EmptyIterSeq2[string, string]()
	}

	return func(yield func(string, string) bool) {
		for name, value := range header.headers {
			if !yield(name, value) {
				break
			}
		}
	}
}

func (header *ReadOnlyHeader) ContentType() ContentType {
	return header.contentType
}

func (header *ReadOnlyHeader) ContentLength() int64 {
	return header.contentLength
}

func (header *ReadOnlyHeader) GetMediaParams(name string) string {
	if header.mediaParams == nil {
		return ""
	}
	return header.mediaParams[name]
}

func (header *ReadOnlyHeader) GetCookie(name string) *Cookie {
	if header.cookies == nil {
		return nil
	}
	return header.cookies[name]
}

func (header *ReadOnlyHeader) Cookies() iter.Seq[Cookie] {
	if header.cookies == nil {
		return utils.EmptyIterSeq[Cookie]()
	}

	return func(yield func(Cookie) bool) {
		for _, cookie := range header.cookies {
			if !yield(*cookie) {
				break
			}
		}
	}
}
