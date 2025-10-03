package llvm

import (
    "os"
    "testing"
)

func TestFindClang_EnvOverride_AMI_CLANG(t *testing.T) {
    oldA, oldC := os.Getenv("AMI_CLANG"), os.Getenv("CLANG")
    defer os.Setenv("AMI_CLANG", oldA); defer os.Setenv("CLANG", oldC)
    _ = os.Setenv("CLANG", "")
    _ = os.Setenv("AMI_CLANG", "/tmp/my-clang")
    got, err := FindClang()
    if err != nil { t.Fatalf("FindClang error: %v", err) }
    if got != "/tmp/my-clang" { t.Fatalf("unexpected path: %s", got) }
}

func TestFindClang_EnvOverride_CLANG(t *testing.T) {
    oldA, oldC := os.Getenv("AMI_CLANG"), os.Getenv("CLANG")
    defer os.Setenv("AMI_CLANG", oldA); defer os.Setenv("CLANG", oldC)
    _ = os.Setenv("AMI_CLANG", "")
    _ = os.Setenv("CLANG", "/tmp/clang")
    got, err := FindClang()
    if err != nil { t.Fatalf("FindClang error: %v", err) }
    if got != "/tmp/clang" { t.Fatalf("unexpected path: %s", got) }
}

