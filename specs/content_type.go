package specs

import (
	"strings"
)

// ContentType defines the content type of a file or data.
//
// It is used to specify the media type of the content being sent or received.
// This is useful in HTTP headers, file uploads, and other scenarios where the type of content
// needs to be communicated clearly.
// The constants defined here are commonly used content types.
// The values are based on the MIME types as defined in RFC 2045 and other relevant standards.
//
// For more information, see: https://www.iana.org/assignments/media-types/media-types.xhtml
const (
	ContentTypeUndefined = ""
	ContentTypeRaw       = "application/octet-stream"
	ContentTypePlain     = "text/plain"
	ContentTypeRichText  = "application/rtf"
	ContentTypeMarkdown  = "text/markdown"

	ContentTypeHTML       = "text/html"
	ContentTypeCSV        = "text/csv"
	ContentTypeCSS        = "text/css"
	ContentTypePDF        = "application/pdf"
	ContentTypeJavaScript = "text/javascript"
	ContentTypeFontTTF    = "font/ttf"

	ContentTypeAVI  = "video/x-msvideo"
	ContentTypeWAV  = "audio/wav"
	ContentTypeMP3  = "audio/mpeg"
	ContentTypeMP4  = "video/mp4"
	ContentTypeMPEG = "video/mpeg"
	ContentTypeMPV  = "video/MPV"
	ContentTypeMKV  = "application/x-matroska"

	ContentTypeAVIF = "image/avif"
	ContentTypeBMP  = "image/bmp"
	ContentTypeGIF  = "image/gif"
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeWEBP = "image/webp"
	ContentTypeSVG  = "image/svg+xml"

	ContentTypeJson           = "application/json"
	ContentTypeXml            = "application/xml"
	ContentTypeMsgpack        = "application/msgpack"
	ContentTypeProtobuf       = "application/x-protobuf"
	ContentTypeForm           = "application/x-www-form-urlencoded"
	ContentTypeMultipart      = "multipart/form-data"
	ContentTypeMultipartMixed = "multipart/mixed"
)

// TODO :: add tests
func MatchContentType(header *Header, expected string) bool {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		return false
	}

	base, _, _ := strings.Cut(contentType, ";")
	mediaType := strings.TrimSpace(strings.ToLower(base))
	return expected == mediaType
}
