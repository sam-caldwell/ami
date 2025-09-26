package scanner

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

func captureUntilEndOfLine(is *Scanner) {
	for is.off < len(is.src) {
		if is.src[is.off] == tok.LexLf {
			break
		}
		is.off++
		is.column++
	}
}
