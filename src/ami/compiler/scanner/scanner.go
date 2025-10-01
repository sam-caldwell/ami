package scanner

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
)

// Scanner performs a minimal lexical scan over a source.File.
type Scanner struct {
	file   *source.File
	offset int
}
