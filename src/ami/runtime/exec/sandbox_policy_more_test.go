package exec

import "testing"

func TestSandboxPolicy_AllowMatrix(t *testing.T) {
    p := SandboxPolicy{AllowFS: true, AllowNet: false, AllowDevice: true}
    if !p.allow("fs") || p.allow("net") || !p.allow("device") { t.Fatalf("policy allow mismatch") }
    if !p.allow("unknown") { t.Fatalf("unknown caps default to allow") }
}

