package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// In JSON+verbose mode, capability/trust violations should stream as diag.v1.
func TestRunBuild_JSONVerbose_CapabilityAndTrust_Diagnostics(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_build", "caps_trust")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    ws := workspace.DefaultWorkspace()
    if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil { t.Fatalf("save ws: %v", err) }

    // Case 1: trust=untrusted with io.read operation -> E_TRUST_VIOLATION
    src1 := `package app
#pragma trust level=untrusted
pipeline P(){ ingress; io.read(); egress }
`
    if err := os.WriteFile(filepath.Join(dir, "src", "t1.ami"), []byte(src1), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out1 bytes.Buffer
    _ = runBuild(&out1, dir, true, true)
    // Scan for E_TRUST_VIOLATION
    foundTrust := false
    s1 := bufio.NewScanner(bytes.NewReader(out1.Bytes()))
    for s1.Scan() {
        var m map[string]any
        if json.Unmarshal(s1.Bytes(), &m) != nil { continue }
        if m["code"] == "E_TRUST_VIOLATION" { foundTrust = true; break }
    }
    if !foundTrust { t.Fatalf("expected E_TRUST_VIOLATION in JSON stream; out=\n%s", out1.String()) }

    // Case 2: no capabilities declared with io.read -> E_CAPABILITY_REQUIRED
    _ = os.Remove(filepath.Join(dir, "src", "t1.ami"))
    src2 := `package app
pipeline P(){ ingress; io.read(); egress }
`
    if err := os.WriteFile(filepath.Join(dir, "src", "t2.ami"), []byte(src2), 0o644); err != nil { t.Fatalf("write: %v", err) }
    var out2 bytes.Buffer
    _ = runBuild(&out2, dir, true, true)
    foundCap := false
    s2 := bufio.NewScanner(bytes.NewReader(out2.Bytes()))
    for s2.Scan() {
        var m map[string]any
        if json.Unmarshal(s2.Bytes(), &m) != nil { continue }
        if m["code"] == "E_CAPABILITY_REQUIRED" { foundCap = true; break }
    }
    if !foundCap { t.Fatalf("expected E_CAPABILITY_REQUIRED in JSON stream; out=\n%s", out2.String()) }
}
