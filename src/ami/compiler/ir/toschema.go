package ir

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// ToSchema converts the module into a schemas.IRV1 for debug output.
func (m Module) ToSchema() sch.IRV1 {
    out := sch.IRV1{Schema: "ir.v1", Package: m.Package, Version: m.Version, File: m.Unit}
    // Build a name→FuncDecl map for parameter introspection if AST present.
    fdecls := map[string]astpkg.FuncDecl{}
    if m.AST != nil {
        for _, d := range m.AST.Decls {
            if fd, ok := d.(astpkg.FuncDecl); ok {
                fdecls[fd.Name] = fd
            }
        }
    }
    for _, fn := range m.Functions {
        irfn := sch.IRFunction{Name: fn.Name, Blocks: []sch.IRBlock{{Label: "entry"}}}
        if fd, ok := fdecls[fn.Name]; ok {
            if len(fd.TypeParams) > 0 {
                for _, tp := range fd.TypeParams {
                    irfn.TypeParams = append(irfn.TypeParams, sch.IRTypeParam{Name: tp.Name, Constraint: tp.Constraint})
                }
            }
            for _, p := range fd.Params {
                dom := ""
                switch {
                case strings.ToLower(p.Type.Name) == "event":
                    dom = "event"
                case p.Type.Name == "State":
                    dom = "state"
                default:
                    dom = "ephemeral"
                }
                own := "borrowed"
                if strings.ToLower(p.Type.Name) == "owned" && len(p.Type.Args) == 1 {
                    own = "owned"
                }
                irfn.Params = append(irfn.Params, sch.IRParam{
                    Name:      p.Name,
                    Type:      typeRefToString(p.Type),
                    Ownership: own,
                    Domain:    dom,
                })
            }
            // Lower a minimal typed IR for imperative subset (scaffold)
            if len(fd.BodyStmts) > 0 {
                var instrs []sch.IRInstr
                var exprStr func(astpkg.Expr) string
                exprStr = func(e astpkg.Expr) string {
                    switch v := e.(type) {
                    case astpkg.Ident:
                        return v.Name
                    case astpkg.BasicLit:
                        return v.Value
                    case astpkg.UnaryExpr:
                        return v.Op + exprStr(v.X)
                    case astpkg.BinaryExpr:
                        return exprStr(v.X) + v.Op + exprStr(v.Y)
                    case astpkg.CallExpr:
                        return "call"
                    case astpkg.ContainerLit:
                        return v.Kind
                    default:
                        return "expr"
                    }
                }
                for _, s := range fd.BodyStmts {
                    switch v := s.(type) {
                    case astpkg.VarDeclStmt:
                        tname := typeRefToString(v.Type)
                        var args []interface{}
                        if v.Name != "" {
                            args = append(args, v.Name)
                        }
                        if tname != "" {
                            args = append(args, tname)
                        }
                        if v.Init != nil {
                            args = append(args, exprStr(v.Init))
                        }
                        instrs = append(instrs, sch.IRInstr{Op: "VAR", Args: args})
                    case astpkg.AssignStmt:
                        instrs = append(instrs, sch.IRInstr{Op: "ASSIGN", Args: []interface{}{exprStr(v.LHS), exprStr(v.RHS)}})
                    case astpkg.ReturnStmt:
                        var rets []interface{}
                        for _, r := range v.Results {
                            rets = append(rets, exprStr(r))
                        }
                        instrs = append(instrs, sch.IRInstr{Op: "RETURN", Args: rets})
                    case astpkg.DeferStmt:
                        instrs = append(instrs, sch.IRInstr{Op: "DEFER"})
                    case astpkg.ExprStmt:
                        instrs = append(instrs, sch.IRInstr{Op: "EXPR", Args: []interface{}{exprStr(v.X)}})
                    }
                }
                irfn.Blocks[0].Instrs = instrs
            }
        }
        out.Functions = append(out.Functions, irfn)
    }
    return out
}
