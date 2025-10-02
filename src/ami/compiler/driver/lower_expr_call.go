package driver

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

func lowerCallExpr(st *lowerState, c *ast.CallExpr) ir.Expr {
    var args []ir.Value
    for _, a := range c.Args {
        if ex, ok := lowerExpr(st, a); ok && ex.Result != nil {
            args = append(args, *ex.Result)
        }
    }
    id := st.newTemp()
    rtype := "any"
    var pSig, rSig []string
    var pNames []string
    if st != nil {
        if st.funcResults != nil {
            if rs, ok := st.funcResults[c.Name]; ok && len(rs) > 0 && rs[0] != "" { rtype = rs[0]; rSig = rs }
        }
        if st.funcParams != nil {
            if ps, ok := st.funcParams[c.Name]; ok { pSig = ps }
        }
        if st.funcParamNames != nil {
            if pn, ok := st.funcParamNames[c.Name]; ok { pNames = pn }
        }
    }
    // Multi-result adaptation: when function signature declares multiple results,
    // synthesize distinct temps and populate Expr.Results instead of single Result.
    if len(rSig) > 1 {
        var results []ir.Value
        for i := range rSig { results = append(results, ir.Value{ID: st.newTemp(), Type: rSig[i]}) }
        return ir.Expr{Op: "call", Callee: c.Name, Args: args, Results: results, ParamTypes: pSig, ParamNames: pNames, ResultTypes: rSig}
    }
    res := &ir.Value{ID: id, Type: rtype}
    return ir.Expr{Op: "call", Callee: c.Name, Args: args, Result: res, ParamTypes: pSig, ParamNames: pNames, ResultTypes: rSig}
}

