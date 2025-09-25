package types

import (
	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"testing"
)

func TestOwnedType_StringAndMapping(t *testing.T) {
	tr := astpkg.TypeRef{Name: "Owned", Args: []astpkg.TypeRef{{Name: "int"}}}
	ty := FromAST(tr)
	if s := ty.String(); s != "Owned<int>" {
		t.Fatalf("Owned mapping string mismatch: %s", s)
	}
}
