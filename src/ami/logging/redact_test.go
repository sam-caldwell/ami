package logging

import (
    "strings"
    "testing"
)

func TestLogger_RedactsConfiguredKeys(t *testing.T) {
    var sb strings.Builder
    lg, err := New(Options{JSON: true, Out: redactBufWriter{&sb}, RedactKeys: []string{"password", "secret"}})
    if err != nil { t.Fatalf("new logger: %v", err) }
    lg.Info("login", map[string]any{"user": "alice", "password": "p@ss", "secret": "token"})
    s := sb.String()
    if !strings.Contains(s, `"password":"[REDACTED]"`) { t.Fatalf("password not redacted: %s", s) }
    if !strings.Contains(s, `"secret":"[REDACTED]"`) { t.Fatalf("secret not redacted: %s", s) }
    if !strings.Contains(s, `"user":"alice"`) { t.Fatalf("user lost: %s", s) }
}

// minimal writer adapter (distinct name to avoid collision with other tests)
type redactBufWriter struct{ b *strings.Builder }
func (w redactBufWriter) Write(p []byte) (int, error) { return w.b.WriteString(string(p)) }
