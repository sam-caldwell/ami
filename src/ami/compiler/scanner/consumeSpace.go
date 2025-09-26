package scanner

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

func consumeSpace(is *Scanner) {
	if is.off < len(is.src) {
		if is.src[is.off] == tok.LexSpace || is.src[is.off] == tok.LexTab {
			is.off++
			is.column++
		}
	}
}
