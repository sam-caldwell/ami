package logging

import "testing"

func TestLevel_Constants(t *testing.T) {
    if LevelInfo != "info" || LevelWarn != "warn" || LevelError != "error" { t.Fatalf("constants mismatch") }
    if LevelTrace != "trace" || LevelDebug != "debug" || LevelFatal != "fatal" { t.Fatalf("constants mismatch") }
}

