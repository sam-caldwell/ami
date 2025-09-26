package root

import scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"

func scannerFor(src string) *scan.Scanner { return scan.New(src) }

