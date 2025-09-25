package tester

import "testing"

func TestRunner_JSONEquality_Pass(t *testing.T) {
    r := New()
    out, err := r.Execute("P", []Case{{Name: "A", Pipeline: "P", InputJSON: `{"x":1}`, ExpectJSON: `{"x":1}`}})
    if err != nil { t.Fatalf("err: %v", err) }
    if len(out) != 1 || out[0].Status != "pass" { t.Fatalf("expected pass, got %+v", out) }
}

func TestRunner_ErrorAssertion_Pass(t *testing.T) {
    r := New()
    out, _ := r.Execute("P", []Case{{Name: "E", Pipeline: "P", InputJSON: `{"error_code":"E_OOPS"}`, ExpectError: "E_OOPS"}})
    if len(out) != 1 || out[0].Status != "pass" { t.Fatalf("expected pass, got %+v", out) }
}

func TestRunner_Timeout_ProducesTimeout(t *testing.T) {
    r := New()
    out, _ := r.Execute("P", []Case{{Name: "T", Pipeline: "P", InputJSON: `{"sleep_ms":20}`, TimeoutMs: 5, ExpectError: "E_TIMEOUT"}})
    if len(out) != 1 || out[0].Status != "pass" { t.Fatalf("expected pass on timeout expectation, got %+v", out) }
}

func TestRunner_Fixture_InvalidMode_Fails(t *testing.T) {
    r := New()
    out, _ := r.Execute("P", []Case{{Name: "F", Pipeline: "P", InputJSON: `{}`, ExpectJSON: `{}`, Fixtures: []Fixture{{Path: "./x", Mode: "bad"}}}})
    if len(out) != 1 || out[0].Status != "fail" { t.Fatalf("expected fail on bad fixture mode, got %+v", out) }
}

