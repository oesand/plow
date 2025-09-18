package prm

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/oesand/plow"
	"github.com/oesand/plow/specs"
)

// FormParam creates a new ParameterProvider for extracting form data from the request body.
// It parses application/x-www-form-urlencoded request bodies into a specs.Query map.
func FormParam(conditions ...Condition[specs.Query]) ParameterProvider[specs.Query] {
	return &formParameter{conditions}
}

type formParameter struct {
	conditions []Condition[specs.Query]
}

func (fp *formParameter) GetParamValue(_ context.Context, req plow.Request) (specs.Query, plow.Response) {
	form, err := plow.ReadForm(req)
	if err != nil {
		return nil, errResponse(err)
	}

	for _, condition := range fp.conditions {
		if err = condition.Validate(form); err != nil {
			return nil, errResponse(err)
		}
	}

	return form, nil
}

// MultipartFormParam creates a new ParameterProvider for extracting multipart form data from the request body.
// It parses multipart/form-data request bodies into a multipart.Reader.
func MultipartFormParam() ParameterProvider[*multipart.Reader] {
	return &multipartFormParameter{}
}

type multipartFormParameter struct{}

func (mfp *multipartFormParameter) GetParamValue(_ context.Context, req plow.Request) (*multipart.Reader, plow.Response) {
	reader, err := plow.MultipartReader(req)
	if err != nil {
		return nil, errResponse(err)
	}

	return reader, nil
}

// JsonParam creates a new ParameterProvider for extracting and parsing JSON data from the request body.
// It parses application/json request bodies into the specified generic type T.
func JsonParam[T any](conditions ...Condition[*T]) ParameterProvider[*T] {
	return &jsonParameter[T]{conditions}
}

type jsonParameter[T any] struct {
	conditions []Condition[*T]
}

func (jp *jsonParameter[T]) GetParamValue(_ context.Context, req plow.Request) (*T, plow.Response) {
	instance, err := plow.ReadJson[T](req)
	if err != nil {
		return nil, errResponse(err)
	}

	for _, condition := range jp.conditions {
		if err = condition.Validate(instance); err != nil {
			return nil, errResponse(err)
		}
	}

	return instance, nil
}

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
