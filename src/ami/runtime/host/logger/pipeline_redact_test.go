package logger

import (
    "strings"
    "testing"
)

func TestPipeline_RedactLogV1Line(t *testing.T) {
    line := []byte("{\"schema\":\"log.v1\",\"timestamp\":\"2025-09-24T17:05:06.123Z\",\"level\":\"info\",\"message\":\"m\",\"fields\":{\"password\":\"x\",\"meta.token\":\"y\"}}\n")
    out, ok := redactLogV1Line(line, []string{"password"}, []string{"meta."})
    if !ok { t.Fatalf("expected redact ok") }
    s := string(out)
    if !containsAll(s, `"password":"[REDACTED]"`, `"meta.token":"[REDACTED]"`) {
        t.Fatalf("expected redactions: %s", s)
    }
}

func containsAll(s string, parts ...string) bool {
    for _, p := range parts {
        if !strings.Contains(s, p) { return false }
    }
    return true
}
