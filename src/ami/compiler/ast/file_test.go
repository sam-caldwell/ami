package ast

import "testing"

func TestFile_Basics(t *testing.T) {
    f := &File{PackageName: "app"}
    if f.PackageName != "app" { t.Fatalf("unexpected package name: %s", f.PackageName) }
}

