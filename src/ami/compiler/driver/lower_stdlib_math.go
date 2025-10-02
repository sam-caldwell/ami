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
    short := name
    if strings.HasPrefix(name, "math.") {
        short = name[len("math."):]
    }
    // Convert args first
    var args []ir.Value
    for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
    // Single-result mappings to llvm.* intrinsics
    one := func(intr string) (ir.Expr, bool) {
        id := st.newTemp()
        res := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: intr, Args: args, Result: res, ResultTypes: []string{"float64"}}, true
    }
    switch short {
    case "FMA": return one("llvm.fma.f64")
    case "Erf": return one("llvm.erf.f64")
    case "Erfc": return one("llvm.erfc.f64")
    case "Abs": return one("llvm.fabs.f64")
    case "Max": return one("llvm.maxnum.f64")
    case "Min": return one("llvm.minnum.f64")
    case "Ceil": return one("llvm.ceil.f64")
    case "Floor": return one("llvm.floor.f64")
    case "Trunc": return one("llvm.trunc.f64")
    case "Round": return one("llvm.round.f64")
    case "RoundToEven": return one("llvm.roundeven.f64")
    case "Exp": return one("llvm.exp.f64")
    case "Expm1": return one("llvm.expm1.f64")
    case "Exp2": return one("llvm.exp2.f64")
    case "Log1p": return one("llvm.log1p.f64")
    case "Log": return one("llvm.log.f64")
    case "Log2": return one("llvm.log2.f64")
    case "Log10": return one("llvm.log10.f64")
    case "Sqrt": return one("llvm.sqrt.f64")
    case "Pow": return one("llvm.pow.f64")
    case "Sin": return one("llvm.sin.f64")
    case "Cos": return one("llvm.cos.f64")
    case "Tan": return one("llvm.tan.f64")
    case "Asin": return one("llvm.asin.f64")
    case "Acos": return one("llvm.acos.f64")
    case "Atan": return one("llvm.atan.f64")
    case "Atan2": return one("llvm.atan2.f64")
    case "Sinh": return one("llvm.sinh.f64")
    case "Cosh": return one("llvm.cosh.f64")
    case "Tanh": return one("llvm.tanh.f64")
    case "Copysign": return one("llvm.copysign.f64")
    case "Nextafter": return one("llvm.nextafter.f64")
    case "Ldexp": return one("llvm.ldexp.f64")
    case "Mod":
        if len(args) >= 2 {
            id := st.newTemp(); res := &ir.Value{ID: id, Type: "float64"}
            return ir.Expr{Op: "mod", Args: []ir.Value{args[0], args[1]}, Result: res, ResultTypes: []string{"float64"}}, true
        }
        return ir.Expr{}, false
    }
    // Multi-result helpers
    switch short {
    case "Sincos":
        // Prefer an aggregate-return runtime helper for portability
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "float64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_sincos", Args: args, Results: res, ResultTypes: []string{"float64", "float64"}}, true
    case "Frexp":
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "int64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_frexp", Args: args, Results: res, ResultTypes: []string{"float64", "int64"}}, true
    case "Modf":
        res := []ir.Value{{ID: st.newTemp(), Type: "float64"}, {ID: st.newTemp(), Type: "float64"}}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_modf", Args: args, Results: res, ResultTypes: []string{"float64", "float64"}}, true
    }
    // NaN/Inf/IsNaN/IsInf/Signbit can be implemented via constants/intrinsics/runtime helpers
    switch short {
    case "NaN":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_nan", Args: nil, Result: r, ResultTypes: []string{"float64"}}, true
    case "Inf":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_inf", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "IsNaN":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_isnan", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "IsInf":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_isinf", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "Signbit":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "bool"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_signbit", Args: args, Result: r, ResultTypes: []string{"bool"}}, true
    case "Remainder":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_remainder", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Pow10":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_pow10", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    }
    // Remaining functions via runtime helpers
    switch short {
    case "Asinh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_asinh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Acosh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_acosh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Atanh":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_atanh", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Cbrt":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_cbrt", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Hypot":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_hypot", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Dim":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_dim", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Logb":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "float64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_logb", Args: args, Result: r, ResultTypes: []string{"float64"}}, true
    case "Ilogb":
        id := st.newTemp(); r := &ir.Value{ID: id, Type: "int64"}
        return ir.Expr{Op: "call", Callee: "ami_rt_math_ilogb", Args: args, Result: r, ResultTypes: []string{"int64"}}, true
    }
    return ir.Expr{}, false
}
