package source

import "testing"

func TestFileSet_AddAndFind(t *testing.T) {
    var s FileSet
    s.AddFile("a.ami", "alpha")
    s.AddFile("b.ami", "beta")
    if s.FileByName("a.ami") == nil { t.Fatalf("expected to find a.ami") }
    if s.FileByName("missing") != nil { t.Fatalf("expected nil for missing file") }
}

