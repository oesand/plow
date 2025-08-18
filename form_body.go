package plow

import (
	"errors"
	"fmt"
	"github.com/oesand/plow/specs"
	"io"
	"mime"
)

// ReadForm reads and parses the request body as a form
// if Content-Type is specs.ContentTypeForm.
func ReadForm(req Request) (specs.Query, error) {
	if req == nil {
		panic("passed nil request")
	}
	body := req.Body()
	if body == nil {
		return nil, errors.New("missing body")
	}

	contentType := req.Header().Get("Content-Type")
	if contentType == "" {
		return nil, errors.New("request Content-Type isn't " + specs.ContentTypeForm)
	}

	ct, _, err := mime.ParseMediaType(contentType)
	if err != nil || ct != specs.ContentTypeForm {
		return nil, errors.New("request Content-Type isn't " + specs.ContentTypeForm)
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, errors.New("failed to read body")
	}

	return specs.ParseQuery(string(b)), nil
}

// FormRequest creates a ClientRequest with the specified HTTP method, URL,
// and form data encoded as application/x-www-form-urlencoded.
// Sets the Content-Type header to specs.ContentTypeForm.
//
// If the method is not specified, defaults to POST. The method must support a request body.
func FormRequest(method specs.HttpMethod, url *specs.Url, form specs.Query) ClientRequest {
	if method == "" {
		method = specs.HttpMethodPost
	} else if !method.IsPostable() {
		panic(fmt.Sprintf("http method '%s' is not postable", method))
	}

	return TextRequest(method, url, specs.ContentTypeForm, form.String())
}
