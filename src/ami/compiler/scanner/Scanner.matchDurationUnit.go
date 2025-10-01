package scanner

// matchDurationUnit returns length of a valid duration unit at position i.
// Recognizes the longest match among: ns, us, ms, h, m, s
func matchDurationUnit(src string, i int) int {
	if i+2 <= len(src) {
		two := src[i : i+2]
		if two == "ns" || two == "us" || two == "ms" {
			return 2
		}
	}
	if i < len(src) {
		switch src[i] {
		case 'h', 'm', 's':
			return 1
		}
	}
	return 0
}
