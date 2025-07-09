package specs

import "strings"

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

	ContentTypeJson      = "application/json"
	ContentTypeXml       = "application/xml"
	ContentTypeMsgpack   = "application/msgpack"
	ContentTypeProtobuf  = "application/x-protobuf"
	ContentTypeForm      = "application/x-www-form-urlencoded"
	ContentTypeMultipart = "multipart/form-data"
)

func IsContentType(header *Header, contentType string) bool {
	return strings.HasPrefix(header.Get("Content-Type"), contentType)
}
