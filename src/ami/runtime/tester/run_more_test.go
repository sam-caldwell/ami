package tester

import (
    "context"
    "testing"
    "time"
)

func TestRun_Identity_More(t *testing.T) {
    in := map[string]any{"k": 1}
    r, err := Run(context.Background(), Options{}, in)
    if err != nil || r.Output["k"].(int) != 1 { t.Fatalf("identity failure: %+v err=%v", r, err) }
}

func TestRun_ErrorCode_More(t *testing.T) {
    _, err := Run(context.Background(), Options{}, map[string]any{"error_code": "E_FAIL"})
    if err == nil { t.Fatalf("expected error") }
}

func TestRun_Timeout_More(t *testing.T) {
    _, err := Run(context.Background(), Options{Timeout: 5 * time.Millisecond}, map[string]any{"sleep_ms": 50})
    if err == nil { t.Fatalf("expected timeout") }
}
