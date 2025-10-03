package llvm

import (
    "os"
    "testing"
)

func TestFindClang_EnvOverrides(t *testing.T) {
    old1, old2 := os.Getenv("AMI_CLANG"), os.Getenv("CLANG")
    defer os.Setenv("AMI_CLANG", old1)
    defer os.Setenv("CLANG", old2)
    os.Setenv("AMI_CLANG", "/tmp/custom/clang1")
    if p, err := FindClang(); err != nil || p != "/tmp/custom/clang1" {
        t.Fatalf("AMI_CLANG override failed: %q %v", p, err)
    }
    os.Setenv("AMI_CLANG", "")
    os.Setenv("CLANG", "/tmp/custom/clang2")
    if p, err := FindClang(); err != nil || p != "/tmp/custom/clang2" {
        t.Fatalf("CLANG override failed: %q %v", p, err)
    }
}

