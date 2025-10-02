package driver

import (
    "fmt"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// foldConst attempts to fold constant subexpressions into literals.
func foldConst(e ast.Expr) ast.Expr {
    switch v := e.(type) {
    case *ast.BinaryExpr:
        x := foldConst(v.X)
        y := foldConst(v.Y)
        // both number literals
        if nx, ok := x.(*ast.NumberLit); ok {
            if ny, ok2 := y.(*ast.NumberLit); ok2 {
                // parse integers with bases: 0x*, 0b*, 0o*, or decimal
                ax, err1 := parseInt(nx.Text)
                ay, err2 := parseInt(ny.Text)
                if err1 == nil && err2 == nil {
                    var r int
                    switch v.Op {
                    case token.Plus: r = ax + ay
                    case token.Minus: r = ax - ay
                    case token.Star: r = ax * ay
                    case token.Slash: if ay != 0 { r = ax / ay } else { return v }
                    default: return v
                    }
                    return &ast.NumberLit{Pos: nx.Pos, Text: fmt.Sprintf("%d", r)}
                }
            }
        }
        // string concatenation
        if sx, ok := x.(*ast.StringLit); ok {
            if sy, ok2 := y.(*ast.StringLit); ok2 && v.Op == token.Plus {
                return &ast.StringLit{Pos: sx.Pos, Value: sx.Value + sy.Value}
            }
        }
        // no fold; but return possibly simplified children
        v.X = x
        v.Y = y
        return v
    default:
        return e
    }
}

