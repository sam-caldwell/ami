package main

import (
    "testing"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func Test_applyConfigSuppress_filters(t *testing.T) {
    ws := &workspace.Workspace{}
    ws.Toolchain.Linter.Suppress = map[string][]string{".": {"X"}}
    ds := applyConfigSuppress("/tmp", ws, []diag.Record{{Code: "X", File: "/tmp/a"}})
    if len(ds) != 0 { t.Fatalf("expected filtered: %+v", ds) }
}

