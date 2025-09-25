package tester

import (
	"bytes"
	"encoding/json"
	lg "github.com/sam-caldwell/ami/src/internal/logger"
	"os"
	"strings"
	"testing"
)

func TestRunner_JSONEquality_Pass(t *testing.T) {
	ResetDefaultKVForTest(t)
	r := New()
	out, err := r.Execute("P", []Case{{Name: "A", Pipeline: "P", InputJSON: `{"x":1}`, ExpectJSON: `{"x":1}`}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) != 1 || out[0].Status != "pass" {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestRunner_ErrorAssertion_Pass(t *testing.T) {
	ResetDefaultKVForTest(t)
	r := New()
	out, _ := r.Execute("P", []Case{{Name: "E", Pipeline: "P", InputJSON: `{"error_code":"E_OOPS"}`, ExpectError: "E_OOPS"}})
	if len(out) != 1 || out[0].Status != "pass" {
		t.Fatalf("expected pass, got %+v", out)
	}
}

func TestRunner_Timeout_ProducesTimeout(t *testing.T) {
	ResetDefaultKVForTest(t)
	r := New()
	out, _ := r.Execute("P", []Case{{Name: "T", Pipeline: "P", InputJSON: `{"sleep_ms":20}`, TimeoutMs: 5, ExpectError: "E_TIMEOUT"}})
	if len(out) != 1 || out[0].Status != "pass" {
		t.Fatalf("expected pass on timeout expectation, got %+v", out)
	}
}

func TestRunner_Fixture_InvalidMode_Fails(t *testing.T) {
	ResetDefaultKVForTest(t)
	r := New()
	out, _ := r.Execute("P", []Case{{Name: "F", Pipeline: "P", InputJSON: `{}`, ExpectJSON: `{}`, Fixtures: []Fixture{{Path: "./x", Mode: "bad"}}}})
	if len(out) != 1 || out[0].Status != "fail" {
		t.Fatalf("expected fail on bad fixture mode, got %+v", out)
	}
}

func TestRunner_EmitsKVStoreMetrics(t *testing.T) {
	ResetDefaultKVForTest(t)
	// Configure logger to JSON and capture stdout
	t.Cleanup(func() { lg.Setup(false, false, false) })
	// We'll use the real Setup each time we need; capture via os.Pipe
	type outpair struct{ s string }
	var out outpair
	{
		// capture inside a closure
		r, w, _ := os.Pipe()
		oldStd := os.Stdout
		os.Stdout = w
		defer func() { os.Stdout = oldStd }()
		// set JSON mode so logger writes diag.v1 JSON to stdout
		lg.Setup(true, true, false)
		// Execute runner with kv directives
		rnr := New()
		_, _ = rnr.Execute("Ptest", []Case{{
			Name:      "K",
			Pipeline:  "Ptest",
			InputJSON: `{"kv_pipeline":"Ptest","kv_node":"NodeA","kv_put_key":"k","kv_put_val":123,"kv_get_key":"k","kv_emit":true}`,
		}})
		w.Close()
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		out.s = buf.String()
	}
	// Search for kvstore.metrics diag
	seen := false
	for _, line := range strings.Split(strings.TrimSpace(out.s), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec map[string]any
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if rec["schema"] == "diag.v1" && rec["message"] == "kvstore.metrics" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("did not observe kvstore.metrics diag output; got:\n%s", out.s)
	}
}

func TestRunner_AutoEmitKV_EmitsMetrics(t *testing.T) {
	ResetDefaultKVForTest(t)
	t.Cleanup(func() { lg.Setup(false, false, false) })
	r, w, _ := os.Pipe()
	oldStd := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStd }()
	lg.Setup(true, true, false)

	rnr := New()
	rnr.EnableAutoEmitKV(true)
	// interact with kv so a store exists
	input, err := BuildKVInput(nil, WithKV("Pauto", "NodeX"), KVPut("ak", 1), KVGet("ak"))
	if err != nil {
		t.Fatalf("compose: %v", err)
	}
	_, _ = rnr.Execute("Pauto", []Case{{Name: "Auto", Pipeline: "Pauto", InputJSON: input}})
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	out := buf.String()
	if !strings.Contains(out, "\"message\":\"kvstore.metrics\"") {
		t.Fatalf("expected kvstore.metrics in output; got:\n%s", out)
	}
}
