package sem

import (
    "fmt"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
    "strings"
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
    // Insert imported package symbols for name resolution (alias or last path segment)
    for _, d := range f.Decls {
        if id, ok := d.(astpkg.ImportDecl); ok {
            name := id.Alias
            if name == "" {
                // derive from path last segment
                parts := strings.Split(id.Path, "/")
                name = parts[len(parts)-1]
            }
            // placeholder type for imported package symbols
            _ = res.Scope.Insert(&types.Object{Kind: types.ObjType, Name: name, Type: types.TPackage})
        }
    }
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            if fd.Name == "_" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IDENT_ILLEGAL", Message: "blank identifier '_' cannot be used as a function name"})
                continue
            }
            // parameters: disallow blank identifier as a parameter name
            for _, p := range fd.Params {
                if p.Name == "_" {
                    res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_PARAM_ILLEGAL", Message: "blank identifier '_' cannot be used as a parameter name"})
                }
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
            // Build a function type from parameter/result TypeRefs
            var ps []types.Type
            for _, p := range fd.Params {
                ps = append(ps, types.FromAST(p.Type))
            }
            var rs []types.Type
            for _, r := range fd.Result {
                rs = append(rs, types.FromAST(r))
            }
            _ = res.Scope.Insert(&types.Object{Kind: types.ObjFunc, Name: fd.Name, Type: types.Function{Params: ps, Results: rs}})
            seen[fd.Name] = true
            funcs[fd.Name] = fd
            // Mutability: enforce AMI semantics (no mut blocks; '*' LHS marker required)
            res.Diagnostics = append(res.Diagnostics, analyzeMutationMarkers(fd)...)
            // Pointer/address prohibitions (2.3.2): '&' not allowed
            res.Diagnostics = append(res.Diagnostics, analyzePointerProhibitions(fd)...)
            // Imperative type checks (2.3): simple assignment type rules (no raw pointers)
            res.Diagnostics = append(res.Diagnostics, analyzeImperativeTypes(fd)...)
            // Call argument type checks with simple generic unification
            res.Diagnostics = append(res.Diagnostics, analyzeCallTypes(fd, funcs)...)
            // Operators: arithmetic/comparison operand compatibility
            res.Diagnostics = append(res.Diagnostics, analyzeOperators(fd)...)
            // Event contracts (1.7): event parameter immutability
            res.Diagnostics = append(res.Diagnostics, analyzeEventContracts(fd)...)
            // State contracts (2.2.14/2.3.5): state param immutability/address-of
            res.Diagnostics = append(res.Diagnostics, analyzeStateContracts(fd)...)
            // Memory domains (2.4): forbid cross-domain references into state
            res.Diagnostics = append(res.Diagnostics, analyzeMemoryDomains(fd)...)
            // Memory model (2.4): ownership & RAII scaffolding
            res.Diagnostics = append(res.Diagnostics, analyzeRAII(fd, funcs)...)
        }
        if ed, ok := d.(astpkg.EnumDecl); ok {
            res.Diagnostics = append(res.Diagnostics, analyzeEnum(ed)...)
        }
        if sd, ok := d.(astpkg.StructDecl); ok {
            res.Diagnostics = append(res.Diagnostics, analyzeStruct(sd)...)
        }
        if id, ok := d.(astpkg.ImportDecl); ok {
            if id.Alias == "_" {
                res.Diagnostics = append(res.Diagnostics, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias"})
            }
        }
        if pd, ok := d.(astpkg.PipelineDecl); ok {
            res.Diagnostics = append(res.Diagnostics, analyzePipeline(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeWorkers(pd, funcs, res.Scope)...)
            res.Diagnostics = append(res.Diagnostics, analyzeEventTypeFlow(pd, funcs)...)
            res.Diagnostics = append(res.Diagnostics, analyzeIOPermissions(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeEdges(pd)...)
            res.Diagnostics = append(res.Diagnostics, analyzeEdgeTypeSafety(pd, funcs)...)
        }
    }
    // Global type checks
    res.Diagnostics = append(res.Diagnostics, analyzeMapTypes(f)...)
    res.Diagnostics = append(res.Diagnostics, analyzeSetTypes(f)...)
    res.Diagnostics = append(res.Diagnostics, analyzeSliceTypes(f)...)
    // Memory domains (2.4 scaffold) across functions
    for _, d := range f.Decls {
        if fd, ok := d.(astpkg.FuncDecl); ok {
            res.Diagnostics = append(res.Diagnostics, analyzeMemoryDomains(fd)...)
        }
    }
    // Cross-pipeline cycle detection (unless cycle pragma present)
    res.Diagnostics = append(res.Diagnostics, analyzeCycles(f)...)
    return res
}

// Check preserves the legacy public API expected by CLI code by returning
// only the diagnostics from AnalyzeFile.
func Check(f *astpkg.File) []diag.Diagnostic {
    return AnalyzeFile(f).Diagnostics
}
