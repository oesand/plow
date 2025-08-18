package prm

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// TODO : Add conditions in body params

// FormParam creates a new ParameterProvider for extracting form data from the request body.
// It parses application/x-www-form-urlencoded request bodies into a specs.Query map.
func FormParam() ParameterProvider[specs.Query] {
	return &formParameter{}
}

type formParameter struct{}

func (fp *formParameter) GetParamValue(_ context.Context, req plow.Request) (specs.Query, plow.Response) {
	body := req.Body()
	if body == nil {
		return nil, ErrorResponse("request body is required")
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, ErrorResponse("failed to read request body")
	}

	formData := specs.ParseQuery(string(bodyBytes))
	return formData, nil
}

// MultipartFormParam creates a new ParameterProvider for extracting multipart form data from the request body.
// It parses multipart/form-data request bodies into a multipart.Reader.
func MultipartFormParam() ParameterProvider[*multipart.Reader] {
	return &multipartFormParameter{}
}

type multipartFormParameter struct{}

func (mfp *multipartFormParameter) GetParamValue(_ context.Context, req plow.Request) (*multipart.Reader, plow.Response) {
	body := req.Body()
	if body == nil {
		return nil, ErrorResponse("request body is required")
	}

	reader, err := plow.MultipartReader(req)
	if err != nil {
		return nil, ErrorResponse("failed to parse multipart form data")
	}

	return reader, nil
}

/*
TODO : fix JsonParam to JsonReader


// JsonParam creates a new ParameterProvider for extracting and parsing JSON data from the request body.
// It parses application/json request bodies into the specified generic type T.
func JsonParam[T any]() ParameterProvider[T] {
	return &jsonParameter[T]{}
}

type jsonParameter[T any] struct{}

func (jp *jsonParameter[T]) GetParamValue(_ context.Context, req plow.Request) (T, plow.Response) {
	body := req.Body()
	if body == nil {
		var zero T
		return zero, ErrorResponse("request body is required")
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		var zero T
		return zero, ErrorResponse("failed to read request body: %s", err)
	}

	var result T
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		var zero T
		return zero, ErrorResponse("failed to parse JSON: %s", err)
	}

	return result, nil
}

*/

// RawBodyParam creates a new ParameterProvider for extracting the raw request body as bytes.
// This is useful when you need to process the body manually or when the content type
// doesn't match the standard form/JSON types.
func RawBodyParam() ParameterProvider[[]byte] {
	return &rawBodyParameter{}
}

type rawBodyParameter struct{}

func (rbp *rawBodyParameter) GetParamValue(_ context.Context, req plow.Request) ([]byte, plow.Response) {
	body := req.Body()
	if body == nil {
		return nil, ErrorResponse("request body is required")
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, ErrorResponse("failed to read request body")
	}

	return bodyBytes, nil
}

// StreamBodyParam creates a new ParameterProvider for accessing the request body as a stream.
// This is useful when you want to process the body as a stream without loading it entirely into memory.
// The body is returned as an io.Reader that can be read incrementally.
func StreamBodyParam() ParameterProvider[io.Reader] {
	return &streamBodyParameter{}
}

type streamBodyParameter struct{}

func (sbp *streamBodyParameter) GetParamValue(_ context.Context, req plow.Request) (io.Reader, plow.Response) {
	body := req.Body()
	if body == nil {
		return nil, ErrorResponse("request body is required")
	}

	return body, nil
}
