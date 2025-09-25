package mod

import "testing"

func TestSelectBackend_KnownSchemes(t *testing.T) {
    if b := selectBackend("./local"); b == nil || b.Name() != "file" {
        t.Fatalf("expected file backend; got %v", b)
    }
    if b := selectBackend("file://./local"); b == nil || b.Name() != "file" {
        t.Fatalf("expected file backend for file://; got %v", b)
    }
    if b := selectBackend("git+ssh://host/org/repo.git#v1.2.3"); b == nil || b.Name() != "git+ssh" {
        t.Fatalf("expected git+ssh backend; got %v", b)
    }
}

