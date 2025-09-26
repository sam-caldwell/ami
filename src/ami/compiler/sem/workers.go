package sem

import (
    "fmt"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "strings"
)

func analyzeWorkers(pd astpkg.PipelineDecl, funcs map[string]astpkg.FuncDecl, scope *types.Scope) []diag.Diagnostic {
    var diags []diag.Diagnostic
    // Check worker references for defined functions or imported package refs
    check := func(ws []astpkg.WorkerRef) {
        for _, w := range ws {
            name := w.Name
            // Allow qualified refs like pkg.Worker
            if dot := strings.IndexByte(name, '.'); dot > 0 {
                pkg := name[:dot]
                if scope != nil {
                    if obj := scope.Lookup(pkg); obj != nil && obj.Type.String() == types.TPackage.String() {
                        // Imported worker reference; cannot validate signature here; skip undefined error.
                        continue
                    }
                }
            }
            fd, ok := funcs[name]
            if !ok {
                // Only flag undefined for explicit factory calls; plain identifiers may
                // refer to imported workers resolved later in lowering.
                if w.Kind == "factory" {
                    diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_WORKER_UNDEFINED", Message: fmt.Sprintf("unknown worker/factory %q", name)})
                }
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
            check(st.Workers)
        }
    }
    for _, st := range pd.ErrorSteps {
        n := strings.ToLower(st.Name)
        switch n {
        case "transform", "fanout":
            check(st.Workers)
        }
    }
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

func isWorkerSignature(fd astpkg.FuncDecl) bool {
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
