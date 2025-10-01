package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStdlibMath maps AMI stdlib math calls to LLVM intrinsics or runtime helpers.
// It returns an IR expression and true when a call was recognized.
func lowerStdlibMath(st *lowerState, c *ast.CallExpr) (ir.Expr, bool) {
    name := c.Name
    if !strings.HasPrefix(name, "math.") { return ir.Expr{}, false }
    // Convert args first
    var args []ir.Value
    for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
    // Single-result mappings to llvm.* intrinsics
    one := func(intr string) (ir.Expr, bool) {
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: intr, Args: args, Result: res, ResultTypes: []string{"float64"}}, true
    }
    switch name {
    case "math.FMA": return one("llvm.fma.f64")
    case "math.Erf": return one("llvm.erf.f64")
    case "math.Erfc": return one("llvm.erfc.f64")
    case "math.Abs": return one("llvm.fabs.f64")
    case "math.Max": return one("llvm.maxnum.f64")
    case "math.Min": return one("llvm.minnum.f64")
    case "math.Ceil": return one("llvm.ceil.f64")
    case "math.Floor": return one("llvm.floor.f64")
    case "math.Trunc": return one("llvm.trunc.f64")
    case "math.Round": return one("llvm.round.f64")
    case "math.RoundToEven": return one("llvm.roundeven.f64")
    case "math.Exp": return one("llvm.exp.f64")
    case "math.Expm1": return one("llvm.expm1.f64")
    case "math.Exp2": return one("llvm.exp2.f64")
    case "math.Log1p": return one("llvm.log1p.f64")
    case "math.Log": return one("llvm.log.f64")
    case "math.Log2": return one("llvm.log2.f64")
    case "math.Log10": return one("llvm.log10.f64")
    case "math.Sqrt": return one("llvm.sqrt.f64")
    case "math.Pow": return one("llvm.pow.f64")
    case "math.Sin": return one("llvm.sin.f64")
    case "math.Cos": return one("llvm.cos.f64")
    case "math.Tan": return one("llvm.tan.f64")
    case "math.Asin": return one("llvm.asin.f64")
    case "math.Acos": return one("llvm.acos.f64")
    case "math.Atan": return one("llvm.atan.f64")
    case "math.Atan2": return one("llvm.atan2.f64")
    case "math.Sinh": return one("llvm.sinh.f64")
    case "math.Cosh": return one("llvm.cosh.f64")
    case "math.Tanh": return one("llvm.tanh.f64")
    case "math.Copysign": return one("llvm.copysign.f64")
    case "math.Nextafter": return one("llvm.nextafter.f64")
    case "math.Ldexp": return one("llvm.ldexp.f64")
    case "math.Mod":
        if len(args) >= 2 {
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "float64"}
            return ir.Expr{Op: "mod", Args: []ir.Value{args[0], args[1]}, Result: res, ResultTypes: []string{"float64"}}, true
        }
        return ir.Expr{}, false
    }
    // Multi-result helpers
    switch name {
    case "math.Sincos":
        // Prefer an aggregate-return runtime helper for portability
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "float64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_sincos", Args: args, Results: res, ResultTypes: []string{"float64", "float64"}}, true
    case "math.Frexp":
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "int64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_frexp", Args: args, Results: res, ResultTypes: []string{"float64", "int64"}}, true
    case "math.Modf":
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "float64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_modf", Args: args, Results: res, ResultTypes: []string{"float64", "float64"}}, true
    }
    // NaN/Inf/IsNaN/IsInf/Signbit can be implemented via constants/intrinsics/runtime helpers
    switch name {
    case "math.NaN":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_nan", Args: nil, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Inf":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_inf", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.IsNaN":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_isnan", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "math.IsInf":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_isinf", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "math.Signbit":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_signbit", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "math.Remainder":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_remainder", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Pow10":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_pow10", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    }
    // Remaining functions via runtime helpers
    switch name {
    case "math.Asinh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_asinh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Acosh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_acosh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Atanh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_atanh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Cbrt":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_cbrt", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Hypot":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_hypot", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Dim":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_dim", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Logb":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_logb", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "math.Ilogb":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_ilogb", Args: args, Result: r, ResultTypes: []string{"int64"}}, true
    }
    return ir.Expr{}, false
}
