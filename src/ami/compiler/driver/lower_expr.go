package driver

import (
    "fmt"
    "strconv"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// lowerExpr lowers an AST expression into an ir.Expr instruction that may yield a result.
// For simple literals/idents, returns an EXPR producing a value of a guessed type.
func lowerExpr(st *lowerState, e ast.Expr) (ir.Expr, bool) {
    // Attempt constant folding first for simpler expressions.
    if fe := foldConst(e); fe != nil { e = fe }
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
    case *ast.UnaryExpr:
        // lower logical not only for now
        if v.Op == token.Bang {
            // lower operand
            ox, ok := lowerExpr(st, v.X)
            if !ok { return ir.Expr{}, false }
            id := st.newTemp()
            res := &ir.Value{ID: id, Type: "i1"}
            var args []ir.Value
            if ox.Result != nil { args = append(args, *ox.Result) }
            return ir.Expr{Op: "not", Args: args, Result: res}, true
        }
        return ir.Expr{}, false
    case *ast.BinaryExpr:
        // simple const-fold for string concatenation
        if v.Op == token.Plus {
            if sx, ok := v.X.(*ast.StringLit); ok {
                if sy, ok2 := v.Y.(*ast.StringLit); ok2 {
                    id := st.newTemp()
                    res := &ir.Value{ID: id, Type: "string"}
                    return ir.Expr{Op: fmt.Sprintf("lit:%q", sx.Value+sy.Value), Result: res}, true
                }
            }
        }
        lx, okx := lowerExpr(st, v.X)
        ly, oky := lowerExpr(st, v.Y)
        if !okx || !oky { return ir.Expr{}, false }
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "any"}
        args := []ir.Value{}
        if lx.Result != nil { args = append(args, *lx.Result) }
        if ly.Result != nil { args = append(args, *ly.Result) }
        return ir.Expr{Op: opName(v.Op), Args: args, Result: res}, true
    case *ast.SliceLit:
        // Lower as a typed container literal with flattened element args.
        id := st.newTemp()
        typ := "slice<" + v.TypeName + ">"
        var args []ir.Value
        for _, el := range v.Elems {
            if ex, ok := lowerExpr(st, el); ok && ex.Result != nil {
                args = append(args, *ex.Result)
            }
        }
        res := &ir.Value{ID: id, Type: typ}
        return ir.Expr{Op: "slice.lit", Args: args, Result: res}, true
    case *ast.SetLit:
        id := st.newTemp()
        typ := "set<" + v.TypeName + ">"
        var args []ir.Value
        for _, el := range v.Elems {
            if ex, ok := lowerExpr(st, el); ok && ex.Result != nil {
                args = append(args, *ex.Result)
            }
        }
        res := &ir.Value{ID: id, Type: typ}
        return ir.Expr{Op: "set.lit", Args: args, Result: res}, true
    case *ast.MapLit:
        id := st.newTemp()
        typ := "map<" + v.KeyType + "," + v.ValType + ">"
        var args []ir.Value
        // Flatten key/value pairs: [k1, v1, k2, v2, ...]
        for _, kv := range v.Elems {
            if kx, ok := lowerExpr(st, kv.Key); ok && kx.Result != nil { args = append(args, *kx.Result) }
            if vx, ok := lowerExpr(st, kv.Val); ok && vx.Result != nil { args = append(args, *vx.Result) }
        }
        res := &ir.Value{ID: id, Type: typ}
        return ir.Expr{Op: "map.lit", Args: args, Result: res}, true
    default:
        return ir.Expr{}, false
    }
}

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

func parseInt(text string) (int, error) {
    // strip optional sign
    if len(text) == 0 { return 0, fmt.Errorf("empty") }
    neg := false
    if text[0] == '-' { neg = true; text = text[1:] }
    base := 10
    if len(text) > 2 && text[0] == '0' {
        switch text[1] {
        case 'x', 'X': base = 16; text = text[2:]
        case 'b', 'B': base = 2;  text = text[2:]
        case 'o', 'O': base = 8;  text = text[2:]
        }
    }
    // remove underscores if any (future-proof)
    clean := make([]rune, 0, len(text))
    for _, r := range text { if r != '_' { clean = append(clean, r) } }
    n, err := strconv.ParseInt(string(clean), base, 64)
    if err != nil { return 0, err }
    if neg { n = -n }
    return int(n), nil
}

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
    res := &ir.Value{ID: id, Type: rtype}
    return ir.Expr{Op: "call", Callee: c.Name, Args: args, Result: res, ParamTypes: pSig, ParamNames: pNames, ResultTypes: rSig}
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
    case token.Percent: return "mod"
    case token.And: return "and"
    case token.Or:  return "or"
    case token.BitXor: return "xor"
    case token.Shl: return "shl"
    case token.Shr: return "shr"
    case token.BitAnd: return "and"
    default:
        return k.String()
    }
}
