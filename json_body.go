package plow

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/oesand/plow/specs"
)

// ReadJson reads a JSON object from a Request or ClientResponse.
// The reqOrResp parameter is the [Request] or [ClientResponse] to read from.
// The T parameter is the type of the JSON object to read.
// The function returns the JSON object and an error if the request or response is not a JSON object.
func ReadJson[T any](reqOrResp any) (*T, error) {
	var header *specs.Header
	var body io.Reader
	switch r := reqOrResp.(type) {
	case Request:
		header = r.Header()
		body = r.Body()
	case ClientResponse:
		header = r.Header()
		body = r.Body()
	default:
		panic("plow: ReadJson support only Request and ClientResponse")
	}

	if body == nil {
		return nil, errors.New("missing body")
	}
	if header.Get("Content-Type") != specs.ContentTypeJson {
		return nil, errors.New("request Content-Type isn't " + specs.ContentTypeJson)
	}
	var res T
	dc := json.NewDecoder(body)
	err := dc.Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

// JsonRequest returns a ClientRequest that can be used to send a JSON request.
// The body parameter is the JSON object to be sent in the request.
func JsonRequest(method specs.HttpMethod, url *specs.Url, body any) (ClientRequest, error) {
	content, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return BufferRequest(method, url, specs.ContentTypeJson, content), nil
}

// JsonResponse returns a MarshallResponse that can be used to send a JSON response.
// The body parameter is the JSON object to be sent in the response.
func JsonResponse(statusCode specs.StatusCode, body any, configure ...func(Response)) MarshallResponse {
	content, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	return &jsonResponse{
		bufferResponse: *newBufferResponse(statusCode, specs.ContentTypeJson, content, configure...),
		instance:       body,
	}
}

type jsonResponse struct {
	bufferResponse
	instance any
}

func (resp *jsonResponse) Instance() any {
	return resp.instance
}
