package sem

import (
    "fmt"
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
    }
    return res
}

