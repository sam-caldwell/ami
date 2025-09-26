package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildImportNode(n astpkg.ImportDecl) sch.ASTNode {
    imp := sch.ASTNode{Kind: "ImportDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    fields := map[string]interface{}{"path": n.Path}
    if n.Alias != "" {
        fields["alias"] = n.Alias
    }
    if n.Constraint != "" {
        fields["constraint"] = n.Constraint
    }
    imp.Fields = fields
    return imp
}

