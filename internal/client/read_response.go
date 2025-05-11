package client

import (
	"bufio"
	"context"
	"github.com/oesand/giglet/internal/parsing"
	"github.com/oesand/giglet/internal/utils/stream"
	"github.com/oesand/giglet/specs"
)

func ReadResponse(ctx context.Context, reader *bufio.Reader, lineLimit int64, totalLimit int64) (*HttpClientResponse, error) {
	select {
	case <-ctx.Done():
		return nil, specs.ErrCancelled
	default:
	}

	line, err := stream.ReadBufferLine(reader, 128)
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
