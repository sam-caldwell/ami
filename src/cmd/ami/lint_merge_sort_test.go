package main

import (
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLint_MergeSort_FieldRequired_And_OrderValid(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "merge_sort")
    srcDir := filepath.Join(dir, "src")
    _ = os.MkdirAll(srcDir, 0o755)
    // Two errors: missing field, invalid order
    content := "package app\npipeline P(){ Ingress().Collect(merge.Sort()).Egress() }\n" +
        "pipeline Q(){ Ingress().Collect(merge.Sort(ts, up)).Egress() }\n"
    _ = os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644)
    ws := workspace.DefaultWorkspace(); ws.Packages[0].Package.Root = "./src"; ws.Toolchain.Linter.Options = []string{}
    _ = ws.Save(filepath.Join(dir, "ami.workspace"))
    setRuleToggles(RuleToggles{StageB: true}); defer setRuleToggles(RuleToggles{})
    var buf bytes.Buffer
    _ = runLint(&buf, dir, true, false, false)
    dec := json.NewDecoder(&buf)
    sawFieldWarn := false
    sawOrderErr := false
    for dec.More() {
        var m map[string]any
        _ = dec.Decode(&m)
        switch m["code"] {
        case "W_MERGE_SORT_NO_FIELD":
            sawFieldWarn = true
        case "E_MERGE_SORT_ORDER_INVALID":
            sawOrderErr = true
        }
    }
    if !sawFieldWarn || !sawOrderErr { t.Fatalf("expected W_MERGE_SORT_NO_FIELD and E_MERGE_SORT_ORDER_INVALID; out=%s", buf.String()) }
}

