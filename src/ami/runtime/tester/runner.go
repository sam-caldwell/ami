package tester

import (
    "encoding/json"
    "errors"
    "reflect"
    "strings"
    "time"
)

// Fixture describes a permitted file path and access mode for a test case.
type Fixture struct {
    Path string
    Mode string // ro|rw
}

// Case describes a Phase 2 runtime test case for a compiled AMI pipeline.
type Case struct {
    Name        string
    Pipeline    string // pipeline name or entry
    InputJSON   string // input payload serialized as JSON
    ExpectJSON  string // expected output payload as JSON
    ExpectError string // expected runtime error code (optional)
    TimeoutMs   int    // per-case timeout
    Fixtures    []Fixture
}

// Result captures outcome for a single case.
type Result struct {
    Name       string
    Status     string // pass|fail|skip
    DurationMs int64
    Error      string // when fail/skip, a description
}

// Runner provides a deterministic runtime executor for AMI pipelines.
// For Phase 2 scaffold, execution is simulated with the following rules:
//  - Unknown pipelines behave as identity functions on the input payload.
//  - Reserved meta keys in input are interpreted by the harness:
//      sleep_ms: integer delay before producing output (for timeout tests)
//      error_code: string causes a runtime error with that code
//  - When ExpectError is set, the produced error must match to pass.
//  - When ExpectJSON is set, the produced output must DeepEqual to pass.
//  - Fixtures are validated for mode only (ro|rw) in this phase.
type Runner struct{}

func New() *Runner { return &Runner{} }

// Execute runs the provided cases against the named pipeline deterministically.
func (r *Runner) Execute(pipeline string, cases []Case) ([]Result, error) {
    res := make([]Result, 0, len(cases))
    for _, c := range cases {
        t0 := time.Now()
        // Validate fixtures (mode only in this phase)
        if err := validateFixtures(c.Fixtures); err != nil {
            res = append(res, Result{Name: c.Name, Status: "fail", Error: err.Error(), DurationMs: time.Since(t0).Milliseconds()})
            continue
        }

        // Parse input JSON (allow empty)
        in, meta, err := parseInput(c.InputJSON)
        if err != nil {
            res = append(res, Result{Name: c.Name, Status: "fail", Error: "invalid input json", DurationMs: time.Since(t0).Milliseconds()})
            continue
        }

        // Handle timeout via simulated sleep_ms
        if meta.sleepMs > 0 {
            if c.TimeoutMs > 0 && meta.sleepMs > c.TimeoutMs {
                // simulate until timeout threshold
                time.Sleep(time.Duration(c.TimeoutMs) * time.Millisecond)
                // then report timeout
                status := "fail"
                msg := "timeout"
                // If expecting timeout specifically, treat as pass
                if strings.EqualFold(c.ExpectError, "E_TIMEOUT") {
                    status = "pass"
                    msg = ""
                }
                res = append(res, Result{Name: c.Name, Status: status, Error: msg, DurationMs: time.Since(t0).Milliseconds()})
                continue
            }
            time.Sleep(time.Duration(meta.sleepMs) * time.Millisecond)
        }

        // Simulate runtime error if requested by input
        if meta.errorCode != "" {
            // Pass only if expected error matches
            if c.ExpectError != "" && c.ExpectError == meta.errorCode {
                res = append(res, Result{Name: c.Name, Status: "pass", DurationMs: time.Since(t0).Milliseconds()})
            } else {
                res = append(res, Result{Name: c.Name, Status: "fail", Error: meta.errorCode, DurationMs: time.Since(t0).Milliseconds()})
            }
            continue
        }

        // Produce output (identity of input without meta keys)
        out := stripMeta(in)

        // Validate expectations
        if c.ExpectError != "" {
            // Expected an error but none occurred
            res = append(res, Result{Name: c.Name, Status: "fail", Error: "no error produced", DurationMs: time.Since(t0).Milliseconds()})
            continue
        }
        if c.ExpectJSON != "" {
            exp, err := decodeJSON(c.ExpectJSON)
            if err != nil {
                res = append(res, Result{Name: c.Name, Status: "fail", Error: "invalid expect_output json", DurationMs: time.Since(t0).Milliseconds()})
                continue
            }
            if !deepEqualJSON(out, exp) {
                res = append(res, Result{Name: c.Name, Status: "fail", Error: "output mismatch", DurationMs: time.Since(t0).Milliseconds()})
                continue
            }
        }
        res = append(res, Result{Name: c.Name, Status: "pass", DurationMs: time.Since(t0).Milliseconds()})
    }
    return res, nil
}

// --- helpers ---

type metaInfo struct{ sleepMs int; errorCode string }

func validateFixtures(fx []Fixture) error {
    for _, f := range fx {
        m := strings.ToLower(strings.TrimSpace(f.Mode))
        if m != "" && m != "ro" && m != "rw" {
            return errors.New("invalid fixture mode")
        }
        if strings.TrimSpace(f.Path) == "" {
            return errors.New("invalid fixture path")
        }
    }
    return nil
}

func parseInput(s string) (any, metaInfo, error) {
    if strings.TrimSpace(s) == "" { return nil, metaInfo{}, nil }
    v, err := decodeJSON(s)
    if err != nil { return nil, metaInfo{}, err }
    // Extract meta keys when object
    mi := metaInfo{}
    if m, ok := v.(map[string]any); ok {
        if n, ok := asInt(m["sleep_ms"]); ok { mi.sleepMs = n }
        if ec, ok := m["error_code"].(string); ok { mi.errorCode = ec }
    }
    return v, mi, nil
}

func asInt(v any) (int, bool) {
    switch t := v.(type) {
    case float64:
        return int(t), true
    case json.Number:
        if i, err := t.Int64(); err == nil { return int(i), true }
    case int:
        return t, true
    }
    return 0, false
}

func decodeJSON(s string) (any, error) {
    dec := json.NewDecoder(strings.NewReader(s))
    dec.UseNumber()
    var v any
    if err := dec.Decode(&v); err != nil { return nil, err }
    return v, nil
}

func stripMeta(v any) any {
    m, ok := v.(map[string]any)
    if !ok { return v }
    out := map[string]any{}
    for k, val := range m {
        kl := strings.ToLower(k)
        if kl == "sleep_ms" || kl == "error_code" { continue }
        out[k] = val
    }
    return out
}

func deepEqualJSON(a, b any) bool {
    return reflect.DeepEqual(a, b)
}
