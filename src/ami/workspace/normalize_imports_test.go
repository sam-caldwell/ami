package workspace

import "testing"

func TestNormalizeImports_Basic(t *testing.T) {
    p := Package{Import: []string{" a ", "a", "b"}}
    NormalizeImports(&p)
    if len(p.Import) != 2 { t.Fatalf("unexpected: %v", p.Import) }
}

