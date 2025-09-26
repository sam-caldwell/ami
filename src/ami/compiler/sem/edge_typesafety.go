package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

// analyzeEdgeTypeSafety validates that declared edge `type=` matches the
// upstream worker output payload type for each step.
func analyzeEdgeTypeSafety(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // helper to get worker result payload type string
    workerOut := func(name string) (string, bool) {
        fd, ok := funcs[name]
        if !ok {
            return "", false
        }
        if len(fd.Result) != 1 {
            return "", false
        }
        r := fd.Result[0]
        // Event<U> or []Event<U>
        if r.Name == "Event" && len(r.Args) == 1 {
            return typeRefToString(fd.Result[0].Args[0]), true
        }
        if r.Name == "Error" && len(r.Args) == 1 {
            return typeRefToString(fd.Result[0].Args[0]), true
        }
        return "", false
    }
    // Compare step i edge type to previous step workers' outputs
    for i := range pd.Steps {
        st := pd.Steps[i]
        spec, ok := parseEdgeSpecFromArgs(st.Args)
        if !ok {
            continue
        }
        // Extract declared type from spec
        var declared string
        switch v := spec.(type) {
        case fifoSpec:
            declared = v.Type
        case lifoSpec:
            declared = v.Type
        case pipeSpec:
            declared = v.Type
        }
        if declared == "" {
            continue
        }
        // ensure previous step exists
        if i == 0 {
            continue
        }
        prev := pd.Steps[i-1]
        // Gather all worker outputs on previous step
        var outs []string
        for _, w := range prev.Workers {
            if t, ok := workerOut(w.Name); ok {
                outs = append(outs, t)
            }
        }
        // If there were no workers on previous step (e.g., Ingress), skip
        if len(outs) == 0 {
            continue
        }
        // All outputs must match declared type
        for _, t := range outs {
            if t != declared {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_TYPE_MISMATCH", Message: "edge type does not match upstream worker output payload"})
                break
            }
        }
    }
    return diags
}

