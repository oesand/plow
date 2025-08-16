package giglet

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/oesand/giglet/specs"
	"io"
	"mime"
	"mime/multipart"
)

// MultipartReader creates a new multipart.Reader for the given request.
// It validates the request's Content-Type header to ensure it is a valid multipart type
// and extracts the boundary parameter required for parsing the multipart body.
func MultipartReader(req Request) (*multipart.Reader, error) {
	if req == nil {
		panic("passed nil request")
	}
	body := req.Body()
	if body == nil {
		return nil, errors.New("missing body")
	}

	contentType := req.Header().Get("Content-Type")
	if contentType == "" {
		return nil, errors.New("request Content-Type isn't " + specs.ContentTypeMultipart)
	}

	ct, params, err := mime.ParseMediaType(contentType)
	if err != nil || !(ct == specs.ContentTypeMultipart || ct == specs.ContentTypeMultipartMixed) {
		return nil, errors.New("request Content-Type isn't " + specs.ContentTypeMultipart)
	}

	boundary, ok := params["boundary"]
	if !ok {
		return nil, errors.New("no multipart boundary param in Content-Type")
	}
	return multipart.NewReader(body, boundary), nil
}

// MultipartWriterFiller is a function type that fills a multipart.Writer with parts.
// It is used to populate the body of a multipart HTTP request.
//
// [multipart.Writer.Close] must not be called by the filler function, as it will be handled by the request itself.
type MultipartWriterFiller func(*multipart.Writer) error

// MultipartRequest creates a new ClientRequest for sending a multipart HTTP request.
// It sets the appropriate Content-Type header with a generated boundary and uses
// the provided filler function to populate the multipart body.
//
// [MultipartWriterFiller] is a function that takes a *multipart.Writer and writes the necessary parts to it.
// [multipart.Writer.Close] must not be called by the filler function, as it will be handled by the request itself.
// Chunked transfer encoding is enabled by default for this request.
//
// If the method is not specified, it defaults to POST.
func MultipartRequest(method specs.HttpMethod, url *specs.Url, filler MultipartWriterFiller) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsPostable() {
		panic(fmt.Sprintf("http method '%s' is not postable", method))
	}

	if filler == nil {
		panic("passed nil multipart form filler")
	}

	req := &multipartRequest{
		clientRequest: *newRequest(method, url),
		filler:        filler,
		boundary:      multipartBoundary(),
	}

	contentType := specs.ContentTypeMultipart + "; boundary=" + req.boundary
	req.Header().Set("Content-Type", contentType)

	specs.WithChunkedEncoding(req.Header())

	return req
}

func multipartBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

type multipartRequest struct {
	clientRequest
	filler   MultipartWriterFiller
	boundary string
}

func (req *multipartRequest) WriteBody(w io.Writer) error {
	writer := multipart.NewWriter(w)
	err := writer.SetBoundary(req.boundary)
	if err != nil {
		return err
	}
	err = req.filler(writer)
	if err != nil {
		return err
	}
	return writer.Close()
}

func (req *multipartRequest) ContentLength() int64 {
	return 0
}
