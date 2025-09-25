package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func writeTempWS(t *testing.T, content string) string {
    t.Helper()
    dir := t.TempDir()
    p := filepath.Join(dir, "ami.workspace")
    if err := os.WriteFile(p, []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    return p
}

func TestValidate_VersionSemver(t *testing.T) {
    // bad semver
    p := writeTempWS(t, "version: not-semver\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected semver validation error for workspace.version")
    }
}

func TestValidate_Concurrency(t *testing.T) {
    // invalid int < 1
    p := writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 0\n    target: ./build\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for concurrency < 1")
    }
    // invalid string (not NUM_CPU)
    p = writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: auto\n    target: ./build\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for invalid concurrency string")
    }
}

func TestValidate_TargetPath(t *testing.T) {
    // absolute path rejected
    p := writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 1\n    target: /abs\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for absolute target path")
    }
    // parent traversal rejected
    p = writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 1\n    target: ../out\n    env: []\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for parent traversal in target path")
    }
}

func TestValidate_EnvPattern(t *testing.T) {
    // invalid os/arch pattern
    p := writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 1\n    target: ./build\n    env:\n      - os: badpattern\n  linker: {}\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for invalid os/arch pattern")
    }
}

func TestValidate_Env_KnownExamplesAndExtensible(t *testing.T) {
    // Known examples plus an extra valid pair to ensure extensibility
    p := writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 2\n    target: ./build\n    env:\n      - os: windows/amd64\n      - os: linux/amd64\n      - os: linux/arm64\n      - os: darwin/amd64\n      - os: darwin/arm64\n      - os: freebsd/riscv64\n  linker: {}\n  linter: {}\npackages: []\n")
    ws, err := Load(p)
    if err != nil { t.Fatalf("unexpected: %v", err) }
    got := make([]string, 0, len(ws.Toolchain.Compiler.Env))
    for _, e := range ws.Toolchain.Compiler.Env { got = append(got, e.OS) }
    want := []string{"windows/amd64", "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "freebsd/riscv64"}
    if len(got) != len(want) { t.Fatalf("env len=%d want %d: %v", len(got), len(want), got) }
    for i := range want {
        if got[i] != want[i] { t.Fatalf("env[%d]=%q want %q", i, got[i], want[i]) }
    }
}

func TestValidate_LinkerLinterObject(t *testing.T) {
    // linker not object
    p := writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 1\n    target: ./build\n    env: []\n  linker: 123\n  linter: {}\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for non-object linker")
    }
    // linter not object
    p = writeTempWS(t, "version: 1.0.0\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: 1\n    target: ./build\n    env: []\n  linker: {}\n  linter: true\npackages: []\n")
    if _, err := Load(p); err == nil {
        t.Fatalf("expected error for non-object linter")
    }
}

func TestValidate_HappyPath(t *testing.T) {
    p := writeTempWS(t, "version: 1.2.3\nproject:\n  name: demo\n  version: 0.0.1\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env:\n      - os: linux/amd64\n      - os: linux/amd64\n      - os: darwin/arm64\n  linker: {}\n  linter: {}\npackages: []\n")
    ws, err := Load(p)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    // dedup env, preserve order
    if len(ws.Toolchain.Compiler.Env) != 2 {
        t.Fatalf("expected deduped env entries, got %d", len(ws.Toolchain.Compiler.Env))
    }
    if ws.ResolveConcurrency() < 1 {
        t.Fatalf("expected ResolveConcurrency >= 1")
    }
}

func TestValidate_PackagesImport_VersionConstraints(t *testing.T) {
    content := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler:
    concurrency: 2
    target: ./build
    env:
      - os: linux/amd64
  linker: {}
  linter: {}
packages:
  - main:
      version: 0.0.1
      root: ./src
      import:
        - github.com/org/repo ^1.2.0
        - github.com/org/repo ~1.2.3
        - github.com/org/repo 1.2.3
        - github.com/org/repo v1.2.3
        - git.example.com/a/b >=1.0.0
        - git.example.com/a/b >1.2.3
        - ./local ==latest
`
    if _, err := Load(writeTempWS(t, content)); err != nil {
        t.Fatalf("unexpected validation error: %v", err)
    }
}

func TestValidate_PackagesImport_InvalidConstraints(t *testing.T) {
    // Too many tokens
    bad := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: 1, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      import:
        - github.com/org/repo ^1.2.3 extra
`
    if _, err := Load(writeTempWS(t, bad)); err == nil {
        t.Fatalf("expected error for too many tokens")
    }
    // Unsupported operator
    bad2 := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: 1, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      import:
        - github.com/org/repo <=1.2.3
`
    if _, err := Load(writeTempWS(t, bad2)); err == nil {
        t.Fatalf("expected error for unsupported operator <=")
    }
    // Invalid semver
    bad3 := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: 1, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      import:
        - github.com/org/repo ^1.2
`
    if _, err := Load(writeTempWS(t, bad3)); err == nil {
        t.Fatalf("expected error for invalid semver ^1.2")
    }
    // Non-string item
    bad4 := `version: 1.0.0
project: { name: demo, version: 0.0.1 }
toolchain: { compiler: { concurrency: 1, target: ./build, env: [] }, linker: {}, linter: {} }
packages:
  - main:
      import:
        - 123
`
    if _, err := Load(writeTempWS(t, bad4)); err == nil {
        t.Fatalf("expected error for non-string import entry")
    }
}
