package workspace

import "testing"

func TestAuditDependencies_NoWorkspace(t *testing.T) {
    // Call with a temp dir lacking workspace; expect error
    dir := t.TempDir()
    if _, err := AuditDependencies(dir); err == nil {
        t.Fatalf("expected error without workspace file")
    }
}

