package driver

import (
    "fmt"
    stdtime "time"
    "strings"

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
    case *ast.ConditionalExpr:
        // Lower condition, then and else branches into a single SSA select.
        cx, okc := lowerExpr(st, v.Cond)
        tx, okt := lowerExpr(st, v.Then)
        ex, oke := lowerExpr(st, v.Else)
        if !okc || !okt || !oke {
            return ir.Expr{}, false
        }
        // Determine a conservative result type: prefer exact match, else 'any'.
        rtype := "any"
        if tx.Result != nil && ex.Result != nil {
            if tx.Result.Type == ex.Result.Type && tx.Result.Type != "" {
                rtype = tx.Result.Type
            }
        }
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: rtype}
        args := []ir.Value{}
        if cx.Result != nil { args = append(args, *cx.Result) }
        if tx.Result != nil { args = append(args, *tx.Result) }
        if ex.Result != nil { args = append(args, *ex.Result) }
        return ir.Expr{Op: "select", Args: args, Result: res}, true
    case *ast.IdentExpr:
        // ident references a previously defined value; use tracked type when known
        typ := "any"
        if st != nil && st.varTypes != nil {
            if t, ok := st.varTypes[v.Name]; ok && t != "" { typ = t }
        }
        res := &ir.Value{ID: v.Name, Type: typ}
        return ir.Expr{Op: "ident", Args: nil, Result: res}, true
    case *ast.SelectorExpr:
        // Recognize enum-like signal selectors; otherwise resolve to a field projection
        // using the receiver's declared type when possible.
        sel := v.Sel
        switch sel {
        case "SIGINT":
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
            return ir.Expr{Op: "lit:2", Result: res}, true
        case "SIGTERM":
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
            return ir.Expr{Op: "lit:15", Result: res}, true
        case "SIGHUP":
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
            return ir.Expr{Op: "lit:1", Result: res}, true
        case "SIGQUIT":
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "int64"}
            return ir.Expr{Op: "lit:3", Result: res}, true
        default:
            if ex, ok := lowerSelectorField(st, v); ok {
                return ex, true
            }
            // Fallback: alias left side
            if lx, ok := lowerExpr(st, v.X); ok && lx.Result != nil {
                res := &ir.Value{ID: lx.Result.ID, Type: lx.Result.Type}
                return ir.Expr{Op: "ident", Result: res}, true
            }
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "any"}
            return ir.Expr{Op: "ident", Result: res}, true
        }
    case *ast.StringLit:
        // strings produce a temp value of type string
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "string"}
        return ir.Expr{Op: fmt.Sprintf("lit:%q", v.Value), Result: res}, true
    case *ast.NumberLit:
        id := st.newTemp()
        lit := v.Text
        typ := "int"
        if strings.ContainsAny(lit, ".eE") { typ = "float64" }
        res := &ir.Value{ID: id, Type: typ}
        return ir.Expr{Op: fmt.Sprintf("lit:%s", lit), Result: res}, true
    case *ast.DurationLit:
        // Parse duration text using Go's parser; represent as int64 nanoseconds literal.
        d, err := stdtime.ParseDuration(v.Text)
        if err != nil { return ir.Expr{}, false }
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: fmt.Sprintf("lit:%d", int64(d)), Result: res}, true
    case *ast.CallExpr:
        // Recognize AMI stdlib intrinsics for lowering
        if ex, ok := lowerStdlibCall(st, v); ok {
            return ex, true
        }
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
        // Set boolean result type for comparisons; otherwise 'any'
        op := opName(v.Op)
        rtype := "any"
        switch op {
        case "eq", "ne", "lt", "le", "gt", "ge":
            rtype = "bool"
        }
        res := &ir.Value{ID: id, Type: rtype}
        args := []ir.Value{}
        if lx.Result != nil { args = append(args, *lx.Result) }
        if ly.Result != nil { args = append(args, *ly.Result) }
        return ir.Expr{Op: op, Args: args, Result: res}, true
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

// Note: helper functions for lowering (foldConst, parseInt, lowerSelectorField,
// flattenSelector, lowerCallExpr, and opName) as well as lowerState.newTemp()
// are defined in separate files to comply with single-declaration-per-file.
