package astjson

import sch "github.com/sam-caldwell/ami/src/schemas"

func buildPackageNode(name string) sch.ASTNode {
    if name == "" {
        return sch.ASTNode{}
    }
    n := sch.ASTNode{Kind: "PackageDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
    n.Fields = map[string]interface{}{"name": name}
    return n
}

