package llvm

import "testing"

// TestTripleFor_EnumerateKnownPairs increases coverage by exercising many mappings.
func TestTripleFor_EnumerateKnownPairs(t *testing.T) {
    cases := map[[2]string]string{
        {"darwin", "amd64"}:   "x86_64-apple-macosx",
        {"darwin", "x86_64"}:  "x86_64-apple-macosx",
        {"darwin", "arm64"}:   "arm64-apple-macosx",
        {"windows", "amd64"}:  "x86_64-pc-windows-msvc",
        {"windows", "x86_64"}: "x86_64-pc-windows-msvc",
        {"windows", "arm64"}:  "aarch64-pc-windows-msvc",
        {"linux", "arm64"}:    "aarch64-unknown-linux-gnu",
        {"linux", "aarch64"}:  "aarch64-unknown-linux-gnu",
        {"linux", "amd64"}:    "x86_64-unknown-linux-gnu",
        {"linux", "x86_64"}:   "x86_64-unknown-linux-gnu",
        {"linux", "riscv64"}:  "riscv64-unknown-linux-gnu",
        {"linux", "arm"}:      "arm-unknown-linux-gnueabihf",
        {"freebsd", "amd64"}:  "x86_64-unknown-freebsd",
        {"freebsd", "x86_64"}: "x86_64-unknown-freebsd",
        {"freebsd", "arm64"}:  "aarch64-unknown-freebsd",
        {"freebsd", "aarch64"}:"aarch64-unknown-freebsd",
        {"openbsd", "amd64"}:  "x86_64-unknown-openbsd",
        {"openbsd", "x86_64"}: "x86_64-unknown-openbsd",
        {"openbsd", "arm64"}:  "aarch64-unknown-openbsd",
        {"openbsd", "aarch64"}:"aarch64-unknown-openbsd",
    }
    for k, want := range cases {
        got := TripleFor(k[0], k[1])
        if got != want {
            t.Fatalf("TripleFor(%s,%s) got %s want %s", k[0], k[1], got, want)
        }
    }
    // unknown defaults
    if got := TripleFor("unknown", "arch"); got != DefaultTriple {
        t.Fatalf("default mismatch: %s", got)
    }
}

