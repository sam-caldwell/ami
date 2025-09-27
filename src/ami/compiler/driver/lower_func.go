package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerFuncDecl lowers a single function declaration into IR with a single entry block.
func lowerFuncDecl(fn *ast.FuncDecl, funcResMap, funcParamMap map[string][]string, funcParamNames map[string][]string) ir.Function {
    var params []ir.Value
    var outResults []ir.Value
    st := &lowerState{varTypes: map[string]string{}, funcResults: funcResMap, funcParams: funcParamMap, funcParamNames: funcParamNames}
    for _, p := range fn.Params {
        params = append(params, ir.Value{ID: p.Name, Type: p.Type})
        if p.Name != "" && p.Type != "" { st.varTypes[p.Name] = p.Type }
    }
    for _, r := range fn.Results {
        outResults = append(outResults, ir.Value{Type: r.Type})
    }
    instrs := lowerBlock(st, fn.Body)
    // collect decorators
    var decos []ir.Decorator
    for _, d := range fn.Decorators {
        var args []string
        for _, e := range d.Args { args = append(args, debugExprText(e)) }
        decos = append(decos, ir.Decorator{Name: d.Name, Args: args})
    }
    blk := ir.Block{Name: "entry", Instr: instrs}
    return ir.Function{Name: fn.Name, Params: params, Results: outResults, Blocks: []ir.Block{blk}, Decorators: decos}
}

// debugExprText mirrors the simple printer used in debug JSON paths.
func debugExprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := debugExprText(v.X)
        if left == "" { left = "?" }
        return left + "." + v.Sel
    case *ast.CallExpr:
        if len(v.Args) > 0 { return v.Name + "(…)" }
        return v.Name + "()"
    case *ast.SliceLit:
        return "slice"
    case *ast.SetLit:
        return "set"
    case *ast.MapLit:
        return "map"
    default:
        return ""
    }
}
