package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

func buildFuncNode(n astpkg.FuncDecl) sch.ASTNode {
    fn := sch.ASTNode{Kind: "FuncDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    fields := map[string]interface{}{"name": n.Name}
    if len(n.TypeParams) > 0 {
        var names []string
        for _, tp := range n.TypeParams { names = append(names, tp.Name) }
        fields["typeParams"] = names
    }
    fn.Fields = fields
    return fn
}
