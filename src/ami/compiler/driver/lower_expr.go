package driver

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// lowerExpr lowers an AST expression into an ir.Expr instruction that may yield a result.
// For simple literals/idents, returns an EXPR producing a value of a guessed type.
func lowerExpr(st *lowerState, e ast.Expr) (ir.Expr, bool) {
    switch v := e.(type) {
    case *ast.IdentExpr:
        // ident references a previously defined value; use tracked type when known
        typ := "any"
        if st != nil && st.varTypes != nil {
            if t, ok := st.varTypes[v.Name]; ok && t != "" { typ = t }
        }
        res := &ir.Value{ID: v.Name, Type: typ}
        return ir.Expr{Op: "ident", Args: nil, Result: res}, true
    case *ast.StringLit:
        // strings produce a temp value of type string
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "string"}
        return ir.Expr{Op: fmt.Sprintf("lit:%q", v.Value), Result: res}, true
    case *ast.NumberLit:
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "int"}
        return ir.Expr{Op: fmt.Sprintf("lit:%s", v.Text), Result: res}, true
    case *ast.CallExpr:
        ex := lowerCallExpr(st, v)
        return ex, true
    case *ast.BinaryExpr:
        lx, okx := lowerExpr(st, v.X)
        ly, oky := lowerExpr(st, v.Y)
        if !okx || !oky { return ir.Expr{}, false }
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "any"}
        args := []ir.Value{}
        if lx.Result != nil { args = append(args, *lx.Result) }
        if ly.Result != nil { args = append(args, *ly.Result) }
        return ir.Expr{Op: opName(v.Op), Args: args, Result: res}, true
    default:
        return ir.Expr{}, false
    }
}

func lowerCallExpr(st *lowerState, c *ast.CallExpr) ir.Expr {
    var args []ir.Value
    for _, a := range c.Args {
        if ex, ok := lowerExpr(st, a); ok && ex.Result != nil {
            args = append(args, *ex.Result)
        }
    }
    id := st.newTemp()
    res := &ir.Value{ID: id, Type: "any"}
    return ir.Expr{Op: "call", Callee: c.Name, Args: args, Result: res}
}

func (s *lowerState) newTemp() string {
    s.temp++
    return fmt.Sprintf("t%d", s.temp)
}

func opName(k token.Kind) string {
    switch k {
    case token.Plus: return "add"
    case token.Minus: return "sub"
    case token.Star: return "mul"
    case token.Slash: return "div"
    default:
        return k.String()
    }
}
