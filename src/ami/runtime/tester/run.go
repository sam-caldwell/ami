package tester

import (
    "context"
    "errors"
    "time"
)

// Run executes a deterministic, side-effect-free simulation:
// - Copies input to output (identity) by default.
// - Recognizes reserved input keys:
//   - sleep_ms (int): delays execution by the given milliseconds.
//   - error_code (string): forces an error with the given code.
func Run(ctx context.Context, opts Options, input map[string]any) (Result, error) {
    start := time.Now()
    // Apply timeout via context
    if opts.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
        defer cancel()
    }
    out := make(map[string]any, len(input))
    for k, v := range input { out[k] = v }
    // Check reserved fields
    if v, ok := input["sleep_ms"]; ok {
        if n, ok := asInt(v); ok && n > 0 {
            t := time.NewTimer(time.Duration(n) * time.Millisecond)
            select {
            case <-ctx.Done():
                t.Stop()
                return Result{Duration: time.Since(start), Output: out, ErrCode: "E_TIMEOUT", ErrMsg: ctx.Err().Error()}, ctx.Err()
            case <-t.C:
            }
        }
    }
    if v, ok := input["error_code"]; ok {
        if s, ok := v.(string); ok && s != "" {
            r := Result{Duration: time.Since(start), ErrCode: s, ErrMsg: s}
            return r, errors.New(s)
        }
    }
    return Result{Output: out, Duration: time.Since(start)}, nil
}

