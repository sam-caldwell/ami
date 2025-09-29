package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeEnums performs compiler checks for enums:
// - Unique members (no duplicates, no blanks) → E_ENUM_MEMBER_DUP / E_ENUM_MEMBER_BLANK
// - Valid literal values when assigned in var declarations with enum type → E_ENUM_ASSIGN_INVALID
func AnalyzeEnums(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // collect enums in this file
    enums := map[string]map[string]struct{}{}
    for _, d := range f.Decls {
        if ed, ok := d.(*ast.EnumDecl); ok {
            set := map[string]struct{}{}
            for _, m := range ed.Members {
                name := strings.TrimSpace(m.Name)
                if name == "" {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ENUM_MEMBER_BLANK", Message: "enum member name cannot be blank", Pos: &diag.Position{Line: m.Pos.Line, Column: m.Pos.Column, Offset: m.Pos.Offset}})
                    continue
                }
                if _, dup := set[name]; dup {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ENUM_MEMBER_DUP", Message: "duplicate enum member: " + name, Pos: &diag.Position{Line: m.Pos.Line, Column: m.Pos.Column, Offset: m.Pos.Offset}})
                } else {
                    set[name] = struct{}{}
                }
            }
            enums[ed.Name] = set
        }
    }
    if len(enums) == 0 { return out }
    // validate enum var assignments in function bodies
    for _, d := range f.Decls {
        fn, ok := d.(*ast.FuncDecl)
        if !ok || fn.Body == nil { continue }
        for _, s := range fn.Body.Stmts {
            if vd, ok := s.(*ast.VarDecl); ok && vd.Type != "" && vd.Init != nil {
                if members, isEnum := enums[vd.Type]; isEnum {
                    if id, ok := vd.Init.(*ast.IdentExpr); ok {
                        if _, ok2 := members[id.Name]; !ok2 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ENUM_ASSIGN_INVALID", Message: "invalid enum assignment: " + id.Name, Pos: &diag.Position{Line: id.Pos.Line, Column: id.Pos.Column, Offset: id.Pos.Offset}, Data: map[string]any{"enum": vd.Type}})
                        }
                    } else {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_ENUM_ASSIGN_INVALID", Message: "invalid enum assignment (must be enum member)", Pos: &diag.Position{Line: vd.Pos.Line, Column: vd.Pos.Column, Offset: vd.Pos.Offset}, Data: map[string]any{"enum": vd.Type}})
                    }
                }
            }
        }
    }
    return out
}

