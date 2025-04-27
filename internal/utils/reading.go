package utils

import (
	"bufio"
)

func ReadBufferLine(reader *bufio.Reader, limit int64) ([]byte, error) {
	var line []byte
	for {
		l, more, err := reader.ReadLine()
		if err != nil {
			return nil, err
		} else if limit > 0 && int64(len(line))+int64(len(l)) > limit {
			return nil, ErrorTooLarge
		} else if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}
