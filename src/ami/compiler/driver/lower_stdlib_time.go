package driver

import (
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// lowerStdlibCall recognizes AMI stdlib calls and lowers them to runtime intrinsics
// or optimized IR forms. It returns (expr, true) when handled.
func lowerStdlibCall(st *lowerState, c *ast.CallExpr) (ir.Expr, bool) {
    if c == nil { return ir.Expr{}, false }
    name := c.Name
    // Normalize alias-qualified call by suffix when possible
    // Supported time intrinsics: time.Sleep(d)
    if strings.HasSuffix(name, ".Sleep") || name == "time.Sleep" {
        // Lower to runtime sleep (milliseconds). Result is void.
        var args []ir.Value
        for _, a := range c.Args { if ex, ok := lowerExpr(st, a); ok && ex.Result != nil { args = append(args, *ex.Result) } }
        return ir.Expr{Op: "call", Callee: "ami_rt_sleep_ms", Args: args}, true
    }
    return ir.Expr{}, false
}

