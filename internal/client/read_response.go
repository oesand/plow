package client

import (
	"bufio"
	"context"
	"errors"
	"github.com/oesand/giglet/internal/parsing"
	"github.com/oesand/giglet/internal/utils"
	"github.com/oesand/giglet/specs"
)

var (
	readResponseOp = specs.GigletOp("client/response")
	ErrorCancelled = &specs.GigletError{Err: errors.New("cancelled")}
)

func ReadResponse(ctx context.Context, reader *bufio.Reader, lineLimit int64, totalLimit int64) (*HttpClientResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ErrorCancelled
	default:
	}

	line, err := utils.ReadBufferLine(reader, 128)
	if err != nil {
		return nil, err
	}

	status, protoMajor, protoMinor, ok := parsing.ParseServerResponseHeadline(line)
	if !ok {
		return nil, specs.NewOpError(readResponseOp, "invalid headline")
	}
	if protoMajor != 1 {
		return nil, specs.NewOpError(readResponseOp, "unsupported http version %d.%d", protoMajor, protoMinor)
	}

	headers, cookies, err := parsing.ParseHeaders(ctx, reader, lineLimit, totalLimit)
	if err != nil {
		return nil, err
	}

	resp := &HttpClientResponse{
		status: status,
		header: specs.NewReadOnlyHeader(headers, cookies),
	}

	return resp, nil
}
