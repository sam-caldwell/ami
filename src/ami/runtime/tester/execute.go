package tester

import (
    kv "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
    mem "github.com/sam-caldwell/ami/src/ami/runtime/memory"
    "strings"
    "sync/atomic"
    "time"
)

// Execute runs the provided cases against the named pipeline deterministically.
func (r *Runner) Execute(pipeline string, cases []Case) ([]Result, error) {
    res := make([]Result, 0, len(cases))
    for _, c := range cases {
        t0 := time.Now()
        // Allocate an ephemeral frame for this case; auto-release at end.
        eph := r.Mem.Alloc(mem.Ephemeral, 1)
        defer eph.Release()
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
        // If input exists, account an Event heap unit; release on case completion.
        if in != nil {
            evh := r.Mem.Alloc(mem.Event, 1)
            defer evh.Release()
        }

        // Optional kvstore interactions for observability tests
        if meta.kvPipeline != "" || meta.kvNode != "" || meta.kvPutKey != "" || meta.kvGetKey != "" || meta.kvEmit {
            pipe := c.Pipeline
            if meta.kvPipeline != "" {
                pipe = meta.kvPipeline
            }
            node := meta.kvNode
            if strings.TrimSpace(node) == "" {
                node = "tester"
            }
            s := kv.Default().Get(pipe, node)
            if meta.kvPutKey != "" {
                s.Put(meta.kvPutKey, meta.kvPutVal)
            }
            if meta.kvGetKey != "" {
                _, _ = s.Get(meta.kvGetKey)
            }
            if meta.kvEmit {
                s.EmitMetrics(pipe, node)
                atomic.AddUint64(&r.kvMetrics, 1)
            }
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
    // Optionally emit kv metrics from all stores after run
    if r.AutoEmitKV {
        infos := snapshotKV()
        for _, inf := range infos {
            s := kv.Default().Get(inf.Pipeline, inf.Node)
            if s != nil {
                s.EmitMetrics(inf.Pipeline, inf.Node)
                atomic.AddUint64(&r.kvMetrics, 1)
            }
        }
    }
    return res, nil
}
