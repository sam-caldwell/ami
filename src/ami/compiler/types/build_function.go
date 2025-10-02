package types

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// BuildFunction constructs a Function type from a FuncDecl's parameter/result type strings.
func BuildFunction(fn *ast.FuncDecl) Function {
    var ps, rs []Type
    if fn != nil {
        for _, p := range fn.Params { ps = append(ps, FromAST(p.Type)) }
        for _, r := range fn.Results { rs = append(rs, FromAST(r.Type)) }
    }
    return Function{Params: ps, Results: rs}
}

