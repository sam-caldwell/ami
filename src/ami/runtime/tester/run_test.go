package tester

import (
    "context"
    "testing"
    "time"
)

func TestRun_Identity(t *testing.T) {
    in := map[string]any{"k": 1}
    r, err := Run(context.Background(), Options{}, in)
    if err != nil { t.Fatalf("unexpected error: %v", err) }
    if r.Output["k"].(int) != 1 { t.Fatalf("bad output: %+v", r.Output) }
}

func TestRun_ErrorCode(t *testing.T) {
    in := map[string]any{"error_code": "E_FAIL"}
    r, err := Run(context.Background(), Options{}, in)
    if err == nil { t.Fatalf("expected error") }
    if r.ErrCode != "E_FAIL" { t.Fatalf("code: %s", r.ErrCode) }
}

func TestRun_Timeout(t *testing.T) {
    in := map[string]any{"sleep_ms": 50}
    _, err := Run(context.Background(), Options{Timeout: 10 * time.Millisecond}, in)
    if err == nil { t.Fatalf("expected timeout") }
}

