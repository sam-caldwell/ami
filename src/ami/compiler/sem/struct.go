package sem

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
)

// analyzeStruct validates struct declarations: non-empty fields, unique names,
// non-blank field names, and presence of a type on each field.
func analyzeStruct(sd astpkg.StructDecl) []diag.Diagnostic {
    var diags []diag.Diagnostic
    if sd.Name == "" {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_NAME", Message: "struct must have a name"})
    }
    if len(sd.Fields) == 0 {
        diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_EMPTY", Message: "struct has no fields"})
        return diags
    }
    seen := map[string]bool{}
    for _, f := range sd.Fields {
        if f.Name == "_" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_BLANK_FIELD", Message: "struct field cannot be '_'"})
        }
        if f.Name == "" {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_NAME", Message: "struct field must have a name"})
        }
        if seen[f.Name] {
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_DUP_FIELD", Message: "duplicate struct field: " + f.Name})
        }
        seen[f.Name] = true
        if f.Type.Name == "" && !f.Type.Ptr && !f.Type.Slice { // no recognizable type
            diags = append(diags, diag.Diagnostic{Level: diag.Error, Code: "E_STRUCT_FIELD_TYPE_INVALID", Message: "struct field missing or invalid type: " + f.Name})
        }
    }
    return diags
}

