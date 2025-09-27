package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Overlapping suppress: parent suppresses W_TODO; child suppresses W_UNKNOWN_IDENT.
// Expect TODO suppressed everywhere under parent, UNKNOWN_IDENT suppressed only under child.
func TestLint_Suppress_Overlap_ParentAndChild(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "suppress_overlap")
    root := filepath.Join(dir, "src")
    sub := filepath.Join(root, "child")
    if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // parent file with TODO and UNKNOWN_IDENT
    if err := os.WriteFile(filepath.Join(root, "a.ami"), []byte("// TODO: parent\nUNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    // child file with TODO and UNKNOWN_IDENT
    if err := os.WriteFile(filepath.Join(sub, "b.ami"), []byte("// TODO: child\nUNKNOWN_IDENT\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    ws.Toolchain.Linter.Options = []string{}
    if ws.Toolchain.Linter.Suppress == nil { ws.Toolchain.Linter.Suppress = map[string][]string{} }
    ws.Toolchain.Linter.Suppress["./src"] = []string{"W_TODO"}
    ws.Toolchain.Linter.Suppress["./src/child"] = []string{"W_UNKNOWN_IDENT"}
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { /* allow */ }
    dec := json.NewDecoder(&buf)
    var sawParentTODO, sawChildTODO bool
    var sawParentUnknown, sawChildUnknown bool
    for dec.More() {
        var m map[string]any
        if derr := dec.Decode(&m); derr != nil { t.Fatalf("json: %v", derr) }
        code, _ := m["code"].(string)
        file, _ := m["file"].(string)
        if code == "W_TODO" {
            if strings.HasSuffix(file, filepath.Join("src", "a.ami")) { sawParentTODO = true }
            if strings.HasSuffix(file, filepath.Join("src", "child", "b.ami")) { sawChildTODO = true }
        }
        if code == "W_UNKNOWN_IDENT" {
            if strings.HasSuffix(file, filepath.Join("src", "a.ami")) { sawParentUnknown = true }
            if strings.HasSuffix(file, filepath.Join("src", "child", "b.ami")) { sawChildUnknown = true }
        }
    }
    if sawParentTODO || sawChildTODO { t.Fatalf("expected W_TODO suppressed under parent; got parent=%v child=%v", sawParentTODO, sawChildTODO) }
    if !sawParentUnknown { t.Fatalf("expected W_UNKNOWN_IDENT present in parent a.ami") }
    if sawChildUnknown { t.Fatalf("expected W_UNKNOWN_IDENT suppressed in child b.ami") }
}

