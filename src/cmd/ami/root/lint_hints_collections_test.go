package root_test

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    rootcmd "github.com/sam-caldwell/ami/src/cmd/ami/root"
    testutil "github.com/sam-caldwell/ami/src/internal/testutil"
)

func TestLint_Hints_MapKeyType_Slice_Warns(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }

    src := "package p\nstruct S { M map<slice<byte>, int> }\n"
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "W_MAP_KEY_TYPE_HINT" { seen = true; break }
    }
    if !seen { t.Fatalf("expected W_MAP_KEY_TYPE_HINT in output; got:\n%s", out) }
}

func TestLint_Hints_SetElemType_Pointer_Warns(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }

    src := "package p\nstruct S { S set<*T> }\n"
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "W_SET_ELEM_TYPE_HINT" { seen = true; break }
    }
    if !seen { t.Fatalf("expected W_SET_ELEM_TYPE_HINT in output; got:\n%s", out) }
}

func TestLint_Hints_SliceArity_Warns(t *testing.T) {
    tmp := t.TempDir()
    t.Setenv("HOME", tmp)
    _, restore := testutil.ChdirToBuildTest(t)
    defer restore()

    ws := `version: 1.0.0
project:
  name: demo
  version: 0.0.1
toolchain:
  compiler: { concurrency: NUM_CPU, target: ./build, env: [] }
  linker: {}
  linter: {}
packages:
  - main: { version: 0.0.1, root: ./src, import: [] }
`
    if err := os.WriteFile("ami.workspace", []byte(ws), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.MkdirAll("src", 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }

    src := "package p\nstruct S { X slice<int, int> }\n"
    if err := os.WriteFile(filepath.Join("src","main.ami"), []byte(src), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    old := os.Args
    os.Args = []string{"ami", "--json", "lint"}
    out := captureStdoutLint(t, func(){ _ = rootcmd.Execute() })
    os.Args = old

    var seen bool
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        var obj map[string]any
        if json.Unmarshal([]byte(sc.Text()), &obj) != nil { continue }
        if obj["schema"] == "diag.v1" && obj["code"] == "W_SLICE_ARITY_HINT" { seen = true; break }
    }
    if !seen { t.Fatalf("expected W_SLICE_ARITY_HINT in output; got:\n%s", out) }
}

