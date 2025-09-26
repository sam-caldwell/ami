package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeEventTypeFlow ensures that the payload type of upstream worker outputs
// matches the Event<T> input payload type of downstream workers for each step
// transition in a pipeline's normal path.
func analyzeEventTypeFlow(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // helper to get worker output payload type string
    workerOut := func(name string) (string, bool) {
        fd, ok := funcs[name]
        if !ok { return "", false }
        // Canonical: (Event<U>, error)
        if len(fd.Result) == 2 {
            r := fd.Result[0]
            if r.Name == "Event" && len(r.Args) == 1 { return typeRefToString(r.Args[0]), true }
            return "", false
        }
        // Legacy single-result Event<U> accepted historically; keep tolerant read here
        if len(fd.Result) == 1 {
            r := fd.Result[0]
            if r.Name == "Event" && len(r.Args) == 1 { return typeRefToString(r.Args[0]), true }
            if r.Name == "Error" && len(r.Args) == 1 { return "", false }
        }
        return "", false
    }
    // helper to get worker input payload type string
    workerIn := func(name string) (string, bool) {
        fd, ok := funcs[name]
        if !ok { return "", false }
        // Canonical: single parameter Event<T>
        if len(fd.Params) >= 1 {
            p := fd.Params[0].Type
            if p.Name == "Event" && len(p.Args) == 1 { return typeRefToString(p.Args[0]), true }
        }
        return "", false
    }
    for i := 1; i < len(pd.Steps); i++ {
        prev := pd.Steps[i-1]
        next := pd.Steps[i]
        var outs []string
        for _, w := range prev.Workers {
            if t, ok := workerOut(w.Name); ok {
                outs = append(outs, t)
            }
        }
        if len(outs) == 0 {
            continue
        }
        var ins []string
        for _, w := range next.Workers {
            if t, ok := workerIn(w.Name); ok {
                ins = append(ins, t)
            }
        }
        // If no downstream worker inputs (e.g., collect/egress), skip
        if len(ins) == 0 {
            continue
        }
        // All combinations must match (with conservative generic compatibility)
        for _, o := range outs {
            for _, in := range ins {
                if o == in {
                    continue
                }
                if isGenericEvent(o) && isGenericEvent(in) {
                    continue
                }
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EVENT_TYPE_FLOW", Message: "event payload type mismatch between upstream worker output and downstream input"})
                // one diag per boundary is enough
                goto nextStep
            }
        }
    nextStep:
        continue
    }
    return diags
}

func isGenericEvent(s string) bool {
    if len(s) < 9 {
        return false
    }
    if !strings.HasPrefix(s, "Event<") || !strings.HasSuffix(s, ">") {
        return false
    }
    inner := s[len("Event<") : len(s)-1]
    if len(inner) == 1 {
        b := inner[0]
        return b >= 'A' && b <= 'Z'
    }
    return false
}
