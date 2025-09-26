package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "strings"
)

// analyzeEdgePipelineTypeSafety verifies that edge.Pipeline(name=X,type=T)
// declarations match the output payload type of pipeline X.
// - Builds a map of pipeline -> output payload type by inspecting the penultimate
//   step workers' result type (the step before egress).
// - If pipeline not found, emits E_EDGE_PIPE_NOT_FOUND.
// - If both declared type and inferred output type are known and do not match,
//   emits E_EDGE_PIPE_TYPE_MISMATCH.
func analyzeEdgePipelineTypeSafety(f *astpkg.File, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Build pipeline -> output payload type map
    outType := map[string]string{}

    // helper to get worker result payload type string
    workerOut := func(name string) (string, bool) {
        fd, ok := funcs[name]
        if !ok { return "", false }
        // Canonical: first result Event<U>
        if len(fd.Result) >= 1 {
            r := fd.Result[0]
            if r.Name == "Event" && len(r.Args) == 1 { return typeRefToString(r.Args[0]), true }
        }
        // Legacy single result Event<U> or Error<E>
        if len(fd.Result) == 1 {
            r := fd.Result[0]
            if r.Name == "Error" && len(r.Args) == 1 { return typeRefToString(r.Args[0]), true }
        }
        return "", false
    }

    // derive pipeline outputs
    for _, d := range f.Decls {
        pd, ok := d.(astpkg.PipelineDecl)
        if !ok || len(pd.Steps) == 0 {
            continue
        }
        // identify penultimate step before egress
        // require last step be egress; otherwise pick last step with workers
        idx := len(pd.Steps) - 1
        if strings.ToLower(pd.Steps[idx].Name) == "egress" {
            idx = idx - 1
        }
        if idx < 0 {
            continue
        }
        st := pd.Steps[idx]

        // Gather worker outputs: prefer attribute-driven worker reference if present,
        // otherwise use positional workers.
        var outs []string
        if wref, ok := st.Attrs["worker"]; ok && strings.TrimSpace(wref) != "" {
            // skip inline func literal text; only handle simple identifiers
            wr := strings.TrimSpace(wref)
            if !strings.HasPrefix(wr, "func") {
                // strip package qualifier if any; lookup full name in funcs map first
                name := wr
                if fd, ok := funcs[name]; ok {
                    if t, ok2 := workerOut(fd.Name); ok2 {
                        outs = append(outs, t)
                    }
                }
            }
        }
        if len(outs) == 0 { // check positional workers
            for _, w := range st.Workers {
                if t, ok := workerOut(w.Name); ok {
                    outs = append(outs, t)
                }
            }
        }
        // consolidate: if all outs equal, record; otherwise leave unknown
        if len(outs) > 0 {
            t := outs[0]
            consistent := true
            for i := 1; i < len(outs); i++ {
                if outs[i] != t {
                    consistent = false
                    break
                }
            }
            if consistent {
                outType[pd.Name] = t
            }
        }
    }

    // Now verify edge.Pipeline uses
    parseFromNode := func(st astpkg.NodeCall) (interface{}, bool) {
        if v := strings.TrimSpace(st.Attrs["in"]); v != "" {
            if spec, ok := parseEdgeSpecFromValue(v); ok {
                return spec, true
            }
        }
        return parseEdgeSpecFromArgs(st.Args)
    }
    for _, d := range f.Decls {
        pd, ok := d.(astpkg.PipelineDecl)
        if !ok {
            continue
        }
        check := func(st astpkg.NodeCall) {
            if spec, ok := parseFromNode(st); ok {
                if p, ok2 := spec.(pipeSpec); ok2 {
                    name := strings.TrimSpace(p.Name)
                    if name == "" {
                        return
                    }
                    // name must refer to a declared pipeline
                    declared, hasType := p.Type, p.Type != ""
                    inferred, found := outType[name]
                    if !found {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_PIPE_NOT_FOUND", Message: "referenced pipeline not found"})
                        return
                    }
                    if hasType && inferred != "" && declared != inferred {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EDGE_PIPE_TYPE_MISMATCH", Message: "edge.Pipeline type does not match referenced pipeline output payload"})
                    }
                }
            }
        }
        for _, st := range pd.Steps {
            check(st)
        }
        for _, st := range pd.ErrorSteps {
            check(st)
        }
    }
    return diags
}
