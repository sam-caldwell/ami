package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_Collections_SliceSetMap_Hints(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "collections")
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package x\nfunc F(){ var a = slice<int>{1}; var b = set<string>{}; var c = map<string,int>{} }\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }
    ws := workspace.DefaultWorkspace()
    ws.Packages[0].Package.Root = "./src"
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save: %v", err) }
    setRuleToggles(RuleToggles{StageB: true})
    defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    var sawSlice, sawSet, sawMap bool
    for dec.More() {
        var m map[string]any
        if err := dec.Decode(&m); err != nil { t.Fatalf("json: %v", err) }
        switch m["code"] {
        case "W_SLICE_ARITY_HINT":
            sawSlice = true
        case "W_SET_EMPTY_HINT":
            sawSet = true
        case "W_MAP_EMPTY_HINT":
            sawMap = true
        }
    }
    if !sawSlice || !sawSet || !sawMap { t.Fatalf("expected collection hints; out=%s", buf.String()) }
}
