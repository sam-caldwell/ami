package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

// analyzeEnum validates enum declarations: non-empty members, unique names,
// valid literal values (if provided), and disallow blank identifier members.
func analyzeEnum(ed astpkg.EnumDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if ed.Name == "" {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_NAME", Message: "enum must have a name"})
    }
    if len(ed.Members) == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_EMPTY", Message: "enum has no members"})
        return diags
    }
    seen := map[string]bool{}
    for _, m := range ed.Members {
        if m.Name == "_" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_BLANK_MEMBER", Message: "enum member cannot be '_'"})
        }
        if seen[m.Name] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_DUP_MEMBER", Message: "duplicate enum member: " + m.Name})
        }
        seen[m.Name] = true
        if m.Value != "" {
            if !(isIntLiteral(m.Value) || isStringLiteral(m.Value)) {
                diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_ENUM_VALUE_INVALID", Message: "enum member value must be integer or string literal: " + m.Name})
            }
        }
    }
    return diags
}

func isIntLiteral(s string) bool {
    if s == "" {
        return false
    }
    i := 0
    if s[0] == '-' {
        if len(s) == 1 {
            return false
        }
        i = 1
    }
    for ; i < len(s); i++ {
        if s[i] < '0' || s[i] > '9' {
            return false
        }
    }
    return true
}

func isStringLiteral(s string) bool {
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
        return true
    }
    return false
}

