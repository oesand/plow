package specs

type ContentType string

const (
	ContentTypeUndefined ContentType = ""
	ContentTypeRaw       ContentType = "application/octet-stream"
	ContentTypePlain     ContentType = "plain/plain"
	ContentTypeRichtext  ContentType = "application/rtf"
	ContentTypeMarkdown  ContentType = "plain/markdown"

	ContentTypeHTML       ContentType = "plain/html"
	ContentTypeCSV        ContentType = "plain/csv"
	ContentTypeCSS        ContentType = "plain/css"
	ContentTypePDF        ContentType = "application/pdf"
	ContentTypeJavaScript ContentType = "plain/javascript"
	ContentTypeFontTTF    ContentType = "font/ttf"

	ContentTypeAVI  ContentType = "video/x-msvideo"
	ContentTypeWAV  ContentType = "audio/wav"
	ContentTypeMP3  ContentType = "audio/mpeg"
	ContentTypeMP4  ContentType = "video/mp4"
	ContentTypeMPEG ContentType = "video/mpeg"
	ContentTypeMPV  ContentType = "video/MPV"
	ContentTypeMKV  ContentType = "application/x-matroska"

	ContentTypeAVIF ContentType = "image/avif"
	ContentTypeBMP  ContentType = "image/bmp"
	ContentTypeGIF  ContentType = "image/gif"
	ContentTypeJPEG ContentType = "image/jpeg"
	ContentTypePNG  ContentType = "image/png"
	ContentTypeWEBP ContentType = "image/webp"
	ContentTypeSVG  ContentType = "image/svg+xml"

	ContentTypeJson      ContentType = "application/json"
	ContentTypeXml       ContentType = "application/xml"
	ContentTypeMsgpack   ContentType = "application/msgpack"
	ContentTypeProtobuf  ContentType = "application/x-protobuf"
	ContentTypeForm      ContentType = "application/x-www-form-urlencoded"
	ContentTypeMultipart ContentType = "multipart/form-data"
)

func (contentType ContentType) IsForm() bool {
	return contentType == ContentTypeForm || contentType == ContentTypeMultipart
}
