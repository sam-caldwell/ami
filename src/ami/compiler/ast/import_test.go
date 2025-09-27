package ast

import "testing"

func TestImportDecl_Basics(t *testing.T) {
    im := &ImportDecl{Path: "alpha"}
    if im.Path != "alpha" { t.Fatalf("path: %q", im.Path) }
}

