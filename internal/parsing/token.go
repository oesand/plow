package parsing

func isTokenChar(r byte) bool {
	switch r {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		switch {
		case r >= '0' && r <= '9':
			return true
		case r >= 'A' && r <= 'Z':
			return true
		case r >= 'a' && r <= 'z':
			return true
		default:
			return false
		}
	}
}
