package server_ops

import (
	"github.com/oesand/plow/specs"
	"io"
)

var closeHeaders = specs.NewHeader(func(header *specs.Header) {
	header.Set("Content-Type", "text/plain; charset=utf-8")
	header.Set("Connection", "close")
})

type ErrorResponse struct {
	Code specs.StatusCode
	Text string
}

func (resp *ErrorResponse) Error() string {
	return "<" + string(resp.Code.Formatted()) + ">: " + resp.Text
}

func (resp *ErrorResponse) WriteTo(writer io.Writer) (int64, error) {
	code := resp.Code
	if code == 0 {
		code = specs.StatusCodeInternalServerError
	}
	size, err := WriteResponseHead(writer, false, code, closeHeaders)
	if err != nil {
		return 0, err
	}
	if resp.Text != "" {
		bodySize, err := writer.Write([]byte(resp.Text))
		if err != nil {
			return 0, err
		}
		size += int64(bodySize)
	}
	return size, nil
}
