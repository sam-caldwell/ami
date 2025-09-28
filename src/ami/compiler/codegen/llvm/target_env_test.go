package llvm

import "testing"

func TestTripleForEnv_MappingsAndDefault(t *testing.T) {
    // known mappings
    if got := TripleForEnv("darwin/arm64"); got != "arm64-apple-macosx" {
        t.Fatalf("darwin/arm64 → %s", got)
    }
    if got := TripleForEnv("linux/amd64"); got != "x86_64-unknown-linux-gnu" {
        t.Fatalf("linux/amd64 → %s", got)
    }
    if got := TripleForEnv("windows/arm64"); got != "aarch64-pc-windows-msvc" {
        t.Fatalf("windows/arm64 → %s", got)
    }
    // default path
    if got := TripleForEnv("unknown/os"); got != DefaultTriple {
        t.Fatalf("unknown default → %s", got)
    }
    if got := TripleForEnv("badformat"); got != DefaultTriple {
        t.Fatalf("bad format default → %s", got)
    }
}

