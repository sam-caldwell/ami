package scanner

func New(src string) *Scanner {
	return &Scanner{
		src:    src,
		line:   1,
		column: 1,
	}
}
