package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

func TestImportAliases_LastSegmentAndAlias(t *testing.T) {
    f := &ast.File{}
    f.Decls = append(f.Decls, &ast.ImportDecl{Path: "alpha/beta"})
    f.Decls = append(f.Decls, &ast.ImportDecl{Path: "gamma", Alias: "g"})
    m := ImportAliases(f)
    if _, ok := m["beta"]; !ok { t.Fatalf("expected last segment alias 'beta'") }
    if _, ok := m["g"]; !ok { t.Fatalf("expected explicit alias 'g'") }
}

