package sem

import (
    "fmt"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "strings"
    srcset "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func analyzeWorkers(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl, scope *types.Scope) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Check worker references for defined functions or imported package refs.
    // Prefer attribute-driven resolution (worker=...), otherwise fall back to positional Workers.
    check := func(st astpkg.NodeCall) {
        pos := &srcset.Position{Line: st.Pos.Line, Column: st.Pos.Column, Offset: st.Pos.Offset}
        // Inline worker literal needs no undefined check
        if st.InlineWorker != nil {
            return
        }
        // Attribute worker= reference
        if name, ok := st.Attrs["worker"]; ok && strings.TrimSpace(name) != "" {
            nm := strings.TrimSpace(name)
            // inline literal (worker=func...) skip undefined check
            if strings.HasPrefix(nm, "func") {
                return
            }
            // Allow qualified refs like pkg.Worker
            name := nm
            // Allow qualified refs like pkg.Worker
            if dot := strings.IndexByte(name, '.'); dot > 0 {
                pkg := name[:dot]
                if scope != nil {
                    if obj := scope.Lookup(pkg); obj != nil && obj.Type.String() == types.TPackage.String() {
                        // Imported worker reference; cannot validate signature here; skip undefined error.
                        return
                    }
                }
            }
            fd, ok := funcs[name]
            if !ok {
                // Only flag undefined for explicit factory calls; plain identifiers may
                // refer to imported workers resolved later in lowering.
                // Here, treat unknown as undefined regardless of kind inference
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: fmt.Sprintf("unknown worker/factory %q", name), Pos: pos})
            } else {
                if !isWorkerSignature(fd) {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name), Pos: pos})
                } else if isLegacyWorkerSignature(fd) {
                    // Transitional notice: explicit State parameter is deprecated; ambient state is preferred
                    diags = append(diags, diag.Diagnostic{Level: diag.Info, Code: "W_WORKER_STATE_PARAM_DEPRECATED", Message: "explicit State parameter is deprecated; state is ambient. Use state.get/set/update/list", Pos: pos})
                }
            }
            return
        }
        // Fallback: positional workers captured by parser
        for _, w := range st.Workers {
            name := w.Name
            if dot := strings.IndexByte(name, '.'); dot > 0 {
                pkg := name[:dot]
                if scope != nil {
                    if obj := scope.Lookup(pkg); obj != nil && obj.Type.String() == types.TPackage.String() {
                        continue
                    }
                }
            }
            if fd, ok := funcs[name]; ok {
                if !isWorkerSignature(fd) {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name), Pos: pos})
                } else if isLegacyWorkerSignature(fd) {
                    diags = append(diags, diag.Diagnostic{Level: diag.Info, Code: "W_WORKER_STATE_PARAM_DEPRECATED", Message: "explicit State parameter is deprecated; state is ambient. Use state.get/set/update/list", Pos: pos})
                }
            } else if w.Kind == "factory" {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: fmt.Sprintf("unknown worker/factory %q", name), Pos: pos})
            }
        }
    }
    for _, st := range pd.Steps { if strings.ToLower(st.Name) == "transform" || strings.ToLower(st.Name) == "fanout" { check(st) } }
    for _, st := range pd.ErrorSteps { if strings.ToLower(st.Name) == "transform" || strings.ToLower(st.Name) == "fanout" { check(st) } }
    // Validate args of transform/fanout for basic shape
    checkArgs := func(args []string) {
        for range args {
            // placeholder for deeper validation
        }
    }
    for _, st := range pd.Steps {
        n := strings.ToLower(st.Name)
        switch n {
        case "transform", "fanout":
            checkArgs(st.Args)
        }
    }
    for _, st := range pd.ErrorSteps {
        n := strings.ToLower(st.Name)
        switch n {
        case "transform", "fanout":
            checkArgs(st.Args)
        }
    }
    return diags
}

func isLegacyWorkerSignature(fd astpkg.FuncDecl) bool {
    // params: (Context, Event<T>, State)
    if len(fd.Params) != 3 {
        return false
    }
    p1 := fd.Params[0].Type
    p2 := fd.Params[1].Type
    p3 := fd.Params[2].Type
    if !(p1.Name == "Context" && !p1.Ptr && !p1.Slice) {
        return false
    }
    if !(p2.Name == "Event" && len(p2.Args) == 1 && !p2.Ptr) {
        return false
    }
    if !(p3.Name == "State") {
        return false
    }
    // results: exactly one of Event<U>, []Event<U>, Error<E>
    if len(fd.Result) != 1 {
        return false
    }
    r := fd.Result[0]
    switch {
    case r.Name == "Event" && len(r.Args) == 1 && !r.Slice:
        return true
    case r.Name == "Event" && len(r.Args) == 1 && r.Slice:
        return true
    case r.Name == "Error" && len(r.Args) == 1:
        return true
    default:
        return false
    }
}

func isCanonicalWorkerSignature(fd astpkg.FuncDecl) bool {
    // params: (Event<T>)
    if len(fd.Params) != 1 {
        return false
    }
    p := fd.Params[0].Type
    if !(p.Name == "Event" && len(p.Args) == 1 && !p.Ptr) {
        return false
    }
    // results: (Event<U>, error)
    if len(fd.Result) != 2 {
        return false
    }
    r1 := fd.Result[0]
    r2 := fd.Result[1]
    if !(r1.Name == "Event" && len(r1.Args) == 1 && !r1.Slice) {
        return false
    }
    if !(strings.ToLower(r2.Name) == "error" && !r2.Slice && !r2.Ptr && len(r2.Args) == 0) {
        return false
    }
    return true
}

func isWorkerSignature(fd astpkg.FuncDecl) bool {
    // Accept canonical or legacy during transition
    if isCanonicalWorkerSignature(fd) {
        return true
    }
    if isLegacyWorkerSignature(fd) {
        return true
    }
    return false
}
