package client

import (
	"bufio"
	"context"
	"github.com/oesand/giglet/internal/catch"
	"github.com/oesand/giglet/internal/parsing"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/specs"
)

func ReadResponse(ctx context.Context, reader *bufio.Reader, lineLimit int64, totalLimit int64) (*HttpClientResponse, error) {
	line, err := stream.ReadBufferLine(reader, lineLimit)
	if err != nil {
		return nil, err
	}

	status, protoMajor, protoMinor, ok := parsing.ParseServerResponseHeadline(line)
	if !ok {
		return nil, specs.NewOpError("parsing", "invalid headline")
	}
	if protoMajor != 1 {
		return nil, specs.NewOpError("parsing", "unsupported http version %d.%d", protoMajor, protoMinor)
	}

	if err = catch.CatchContextCancel(ctx); err != nil {
		return nil, err
	}

	header, err := parsing.ParseHeaders(ctx, reader, lineLimit, totalLimit)
	if err != nil {
		return nil, err
	}

	resp := &HttpClientResponse{
		status: status,
		header: header,
	}

	return resp, nil
}
