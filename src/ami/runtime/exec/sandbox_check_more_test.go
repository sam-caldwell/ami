package exec

import "testing"

func TestSandboxCheck_DeniedReturnsError(t *testing.T) {
    err := sandboxCheck(SandboxPolicy{AllowDevice: false}, "device")
    if err == nil { t.Fatalf("expected sandbox denial error") }
}

