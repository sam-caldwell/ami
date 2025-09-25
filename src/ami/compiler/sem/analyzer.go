package sem

import (
    "fmt"
    "strings"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

type Result struct {
    Scope       *types.Scope
    Diagnostics []diag.Diagnostic
}

// AnalyzeFile performs minimal semantic analysis:
// - Build a top-level scope
// - Detect duplicate function declarations
// - Validate basic pipeline semantics (ingress→...→egress)
func AnalyzeFile(f *astpkg.File) Result {
    res := Result{Scope: types.NewScope(nil)}
    seen := map[string]bool{}
    // collect function names for worker resolution
    funcs := map[string]astpkg.FuncDecl{}
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            if fd.Name == "_" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IDENT_ILLEGAL", Message: "blank identifier '_' cannot be used as a function name"})
                continue
            }
            if seen[fd.Name] {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{
                    Level:   diag.Error,
                    Code:    "E_DUP_FUNC",
                    Message: fmt.Sprintf("duplicate function %q", fd.Name),
                    File:    "",
                })
                continue
            }
            _ = res.Scope.Insert(&types.Object{Kind: types.ObjFunc, Name: fd.Name, Type: types.TInvalid})
            seen[fd.Name] = true
            funcs[fd.Name] = fd
        }
        if id, ok := d.(astpkg.ImportDecl); ok {
            if id.Alias == "_" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias"})
            }
        }
        if pd, ok := d.(astpkg.PipelineDecl); ok {
            res.Diagnostics = append(res.Diagnostics, analyzePipeline(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeWorkers(pd, funcs)...)
        }
    }
    return res
}

func analyzePipeline(pd astpkg.PipelineDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Basic pipeline shape checks
    if len(pd.Steps) == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_EMPTY", Message: fmt.Sprintf("pipeline %q has no steps", pd.Name)})
        return diags
    }
    allowed := map[string]bool{"ingress":true, "transform":true, "fanout":true, "collect":true, "egress":true}
    ingressCount := 0
    egressCount := 0
    for i, step := range pd.Steps {
        name := strings.ToLower(step.Name)
        if !allowed[name] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: fmt.Sprintf("unknown node %q", step.Name)})
            continue
        }
        switch name {
        case "ingress":
            ingressCount++
            if i != 0 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_INGRESS_POSITION", Message: fmt.Sprintf("pipeline %q: ingress must be first", pd.Name)})
            }
        case "egress":
            egressCount++
            if i != len(pd.Steps)-1 {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_EGRESS_POSITION", Message: fmt.Sprintf("pipeline %q: egress must be last", pd.Name)})
            }
        }
    }
    if ingressCount == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_START_INGRESS", Message: fmt.Sprintf("pipeline %q must start with ingress", pd.Name)})
    }
    if egressCount == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_PIPELINE_END_EGRESS", Message: fmt.Sprintf("pipeline %q must end with egress", pd.Name)})
    }
    if ingressCount > 1 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_INGRESS", Message: fmt.Sprintf("pipeline %q has multiple ingress nodes", pd.Name)}) }
    if egressCount > 1 { diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_DUP_EGRESS", Message: fmt.Sprintf("pipeline %q has multiple egress nodes", pd.Name)}) }

    // Error pipeline semantics
    if len(pd.ErrorSteps) > 0 {
        if strings.ToLower(pd.ErrorSteps[0].Name) == "ingress" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_START_INVALID", Message: fmt.Sprintf("pipeline %q error path cannot start with ingress", pd.Name)})
        }
        if strings.ToLower(pd.ErrorSteps[len(pd.ErrorSteps)-1].Name) != "egress" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ERRPIPE_END_EGRESS", Message: fmt.Sprintf("pipeline %q error path must end with egress", pd.Name)})
        }
        for _, st := range pd.ErrorSteps {
            nm := strings.ToLower(st.Name)
            if !allowed[nm] {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: fmt.Sprintf("unknown node %q in error path", st.Name)})
            }
        }
    }
    return diags
}

// analyzeWorkers ensures worker/factory references in pipeline steps resolve
// to known top-level function declarations. It scans step args heuristically:
// - IDENT or IDENT(arg,...) → resolves to IDENT
// Applies to Transform and FanOut nodes.
func analyzeWorkers(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    checkArgs := func(args []string) {
        for _, a := range args {
            name := a
            hasCall := false
            if i := strings.IndexRune(a, '('); i >= 0 { name = strings.TrimSpace(a[:i]); hasCall = true }
            // simple identifier extract: letters/_ followed by letters/digits/_
            if name == "" { continue }
            // skip placeholders like "cfg" or literals
            if name == "cfg" { continue }
            // Only enforce for explicit calls (factory invocations) or names starting with New
            if !(hasCall || strings.HasPrefix(name, "New")) {
                // if bare name, only check signature if function exists; otherwise skip
                if fd, ok := funcs[name]; ok {
                    if !isWorkerSignature(fd) {
                        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name)})
                    }
                }
                continue
            }
            fd, ok := funcs[name]
            if !ok {
                // allow blank identifier '_' to pass worker ref check as sink
                if name == "_" { continue }
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: fmt.Sprintf("unknown worker/factory %q", name)})
            } else {
                if !isWorkerSignature(fd) {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_SIGNATURE", Message: fmt.Sprintf("worker %q has invalid signature", name)})
                }
            }
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

func isWorkerSignature(fd astpkg.FuncDecl) bool {
    // params: (Context, Event<T>, *State)
    if len(fd.Params) != 3 { return false }
    p1 := fd.Params[0].Type
    p2 := fd.Params[1].Type
    p3 := fd.Params[2].Type
    if !(p1.Name == "Context" && !p1.Ptr && !p1.Slice) { return false }
    if !(p2.Name == "Event" && len(p2.Args) == 1 && !p2.Ptr) { return false }
    if !(p3.Name == "State" && p3.Ptr) { return false }
    // results: exactly one of Event<U>, []Event<U>, Error<E>, Drop/Ack
    if len(fd.Result) != 1 { return false }
    r := fd.Result[0]
    switch {
    case r.Name == "Event" && len(r.Args) == 1 && !r.Slice:
        return true
    case r.Name == "Event" && len(r.Args) == 1 && r.Slice:
        return true
    case r.Name == "Error" && len(r.Args) == 1:
        return true
    case r.Name == "Drop" && len(r.Args) == 0:
        return true
    case r.Name == "Ack" && len(r.Args) == 0:
        return true
    default:
        return false
    }
}
