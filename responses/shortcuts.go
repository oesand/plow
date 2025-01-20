package responses

import (
	"github.com/oesand/giglet"
	"github.com/oesand/giglet/internal"
	"github.com/oesand/giglet/specs"
	"io"
	"strconv"
)

func Redirect(url string) giglet.Response {
	resp := &HeaderResponse{}
	resp.SetStatusCode(specs.StatusCodeTemporaryRedirect)
	resp.Header().Set("Location", url)
	return resp
}

func PermanentRedirect(url string) giglet.Response {
	resp := &HeaderResponse{}
	resp.SetStatusCode(specs.StatusCodePermanentRedirect)
	resp.Header().Set("Location", url)
	return resp
}

func EmptyResponse(contentType specs.ContentType, configure ...func(response giglet.Response)) giglet.Response {
	resp := &HeaderResponse{}

	if contentType != specs.ContentTypeUndefined {
		resp.Header().Set("Content-Type", string(contentType))
	}

	for _, conf := range configure {
		conf(resp)
	}

	return resp
}

func TextResponse(text string, contentType specs.ContentType, configure ...func(response giglet.Response)) giglet.Response {
	resp := &bufferResponse{buffer: internal.StringToBuffer(text)}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypePlain
	}
	resp.Header().Set("Content-Length", strconv.Itoa(len(text)))
	resp.Header().Set("Content-Type", string(contentType))

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

func BufferResponse(buffer []byte, contentType specs.ContentType, configure ...func(response giglet.Response)) giglet.Response {
	resp := &bufferResponse{buffer: buffer}

	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	resp.Header().Set("Content-Length", strconv.Itoa(len(buffer)))
	resp.Header().Set("Content-Type", string(contentType))

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

type bufferResponse struct {
	HeaderResponse
	buffer []byte
}

func (resp *bufferResponse) WriteBody(writer io.Writer) {
	writer.Write(resp.buffer)
}

func StreamResponse(stream io.Reader, size uint64, contentType specs.ContentType, configure ...func(response giglet.Response)) giglet.Response {
	resp := &streamResponse{stream: stream}
	if contentType == specs.ContentTypeUndefined {
		contentType = specs.ContentTypeRaw
	}
	if size > 0 {
		resp.Header().Set("Content-Length", strconv.FormatUint(size, 10))
	}
	resp.Header().Set("Content-Type", string(contentType))

	for _, conf := range configure {
		conf(&resp.HeaderResponse)
	}

	return resp
}

type streamResponse struct {
	HeaderResponse
	stream io.Reader
}

func (resp *streamResponse) WriteBody(writer io.Writer) {
	io.Copy(writer, resp.stream)
}
