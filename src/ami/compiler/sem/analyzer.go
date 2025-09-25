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
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
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
        }
        if pd, ok := d.(astpkg.PipelineDecl); ok {
            // Basic pipeline shape checks
            if len(pd.Steps) == 0 {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{
                    Level:   diag.Error,
                    Code:    "E_PIPELINE_EMPTY",
                    Message: fmt.Sprintf("pipeline %q has no steps", pd.Name),
                    File:    "",
                })
                continue
            }
            // Normalize node names for comparison
            first := strings.ToLower(pd.Steps[0].Name)
            last := strings.ToLower(pd.Steps[len(pd.Steps)-1].Name)
            if first != "ingress" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{
                    Level:   diag.Error,
                    Code:    "E_PIPELINE_START_INGRESS",
                    Message: fmt.Sprintf("pipeline %q must start with ingress", pd.Name),
                    File:    "",
                })
            }
            if last != "egress" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{
                    Level:   diag.Error,
                    Code:    "E_PIPELINE_END_EGRESS",
                    Message: fmt.Sprintf("pipeline %q must end with egress", pd.Name),
                    File:    "",
                })
            }
            // Unknown node kinds produce diagnostics
            allowed := map[string]bool{"ingress":true, "transform":true, "fanout":true, "collect":true, "egress":true}
            for _, step := range pd.Steps {
                name := strings.ToLower(step.Name)
                if !allowed[name] {
                    res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_UNKNOWN_NODE", Message: fmt.Sprintf("unknown node %q", step.Name)})
                }
            }
        }
    }
    return res
}
