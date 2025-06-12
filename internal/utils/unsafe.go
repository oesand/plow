package utils

import "unsafe"

func StringToBuffer(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BufferToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
