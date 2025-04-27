package server

import (
	"github.com/oesand/giglet/internal/writing"
	"github.com/oesand/giglet/specs"
	"io"
)

var CloseHeaders = specs.NewHeader(func(header *specs.Header) {
	header.Set("Content-Type", "text/plain; charset=utf-8")
	header.Set("Connection", "close")
})

type ErrorResponse struct {
	Code specs.StatusCode
	Text string
}

func (resp *ErrorResponse) Error() string {
	return string(resp.Code.Detail()) + ": " + resp.Text
}

func (resp *ErrorResponse) Write(writer io.Writer) error {
	_, err := writing.WriteResponseHead(writer, false, resp.Code, CloseHeaders)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(resp.Text))
	return err
}
