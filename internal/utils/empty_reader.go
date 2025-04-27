package utils

import "io"

var EmptyReader = emptyReader{}

type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error)         { return 0, io.EOF }
func (emptyReader) Close() error                     { return nil }
func (emptyReader) WriteTo(io.Writer) (int64, error) { return 0, nil }
