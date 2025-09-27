package workspace

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
)

func TestAuditDependencies_ReportsMissingAndUnsatisfied(t *testing.T) {
    dir := filepath.Join("build", "test", "audit", "ws1")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // workspace with remote imports
    wsY := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.2.0', 'modB >= 1.0.0', 'modC 1.2.3', 'modD <= 1.2.3' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsY, 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    // sum with modA@1.2.3 valid, modB@0.9.0 only
    sum := Manifest{Schema: "ami.sum/v1"}
    sum.Set("modA", "1.2.3", "shaA")
    sum.Set("modB", "0.9.0", "shaB")
    // write ami.sum
    if err := sum.Save(filepath.Join(dir, "ami.sum")); err != nil { t.Fatalf("save sum: %v", err) }

    rep, err := AuditDependencies(dir)
    if err != nil { t.Fatalf("audit: %v", err) }
    if !rep.SumFound { t.Fatalf("expected sum found") }
    // modC missing in sum, modB unsatisfied, modA ok
    if len(rep.MissingInSum) != 1 || rep.MissingInSum[0] != "modC" { t.Fatalf("missingInSum: %v", rep.MissingInSum) }
    if len(rep.Unsatisfied) != 1 || rep.Unsatisfied[0] != "modB" { t.Fatalf("unsatisfied: %v", rep.Unsatisfied) }
    // parse error for modD <= 1.2.3
    if len(rep.ParseErrors) == 0 { b, _ := json.Marshal(rep); t.Fatalf("expected parse errors; rep=%s", string(b)) }
}

func TestAuditDependencies_MissingSum_AllMissing(t *testing.T) {
    dir := filepath.Join("build", "test", "audit", "ws2")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    wsY := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: [ 'modA ^1.0.0', 'modB 2.0.0' ]\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsY, 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    rep, err := AuditDependencies(dir)
    if err != nil { t.Fatalf("audit: %v", err) }
    if rep.SumFound { t.Fatalf("sum should be missing") }
    if len(rep.MissingInSum) != 2 { t.Fatalf("expected both missing; got %v", rep.MissingInSum) }
    if len(rep.Unsatisfied) != 0 { t.Fatalf("unsatisfied should be empty; got %v", rep.Unsatisfied) }
}

