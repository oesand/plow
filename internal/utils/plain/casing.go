package plain

func TitleCase(content string) string {
	return string(TitleCaseBytes([]byte(content)))
}

func TitleCaseBytes(input []byte) []byte {
	output := make([]byte, len(input))
	capNext := true // capitalize first letter or anything after space

	for i, b := range input {
		if 'a' <= b && b <= 'z' && capNext {
			b -= 32 // to upper
		} else if 'A' <= b && b <= 'Z' && !capNext {
			b += 32 // to lowercase
		}

		output[i] = b
		switch b {
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0, '-', '_':
			capNext = true
		default:
			capNext = false
		}
	}

	return output
}
