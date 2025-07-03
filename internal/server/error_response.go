package server

import (
	"github.com/oesand/giglet/specs"
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

func (resp *ErrorResponse) WriteTo(writer io.Writer) error {
	_, err := WriteResponseHead(writer, false, resp.Code, closeHeaders)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(resp.Text))
	return err
}
