package logging

import (
    "strings"
    "testing"
)

func TestLogger_FilterAllowAndDenyAndPrefix(t *testing.T) {
    var sb strings.Builder
    lg, err := New(Options{JSON: true, Out: bufWriter{&sb},
        FilterAllowKeys: []string{"user", "password", "secret", "meta.token"},
        FilterDenyKeys:  []string{"secret"},
        RedactKeys:      []string{"password"},
        RedactPrefixes:  []string{"meta."},
    })
    if err != nil { t.Fatalf("new logger: %v", err) }
    lg.Info("login", map[string]any{"user":"alice","password":"p@ss","secret":"hide","other":"drop","meta.token":"abc"})
    s := sb.String()
    if strings.Contains(s, `"other"`) { t.Fatalf("allowlist failed to drop unknown key: %s", s) }
    if strings.Contains(s, `"secret"`) { t.Fatalf("denylist failed to drop key: %s", s) }
    if !strings.Contains(s, `"password":"[REDACTED]"`) { t.Fatalf("redact didn't apply: %s", s) }
    if !strings.Contains(s, `"meta.token":"[REDACTED]"`) { t.Fatalf("prefix redact didn't apply: %s", s) }
}

