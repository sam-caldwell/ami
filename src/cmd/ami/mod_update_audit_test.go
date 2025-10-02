package main

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func Test_embedAudit(t *testing.T) {
    if embedAudit(workspace.AuditReport{}) == nil { t.Fatal("nil") }
}
