package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// ToSchemaAST converts an internal AST File into schemas.ASTV1 with
// a richer structural tree suitable for debug artifacts.
func ToSchemaAST(file *astpkg.File, filePath string) sch.ASTV1 {
    root := sch.ASTNode{Kind: "File", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}

    // package declaration
    if file.Package != "" {
        pkgNode := sch.ASTNode{Kind: "PackageDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
        pkgNode.Fields = map[string]interface{}{"name": file.Package}
        root.Children = append(root.Children, pkgNode)
    }

    // declarations in order
    for _, d := range file.Decls {
        switch n := d.(type) {
        case astpkg.ImportDecl:
            imp := sch.ASTNode{Kind: "ImportDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            fields := map[string]interface{}{"path": n.Path}
            if n.Alias != "" { fields["alias"] = n.Alias }
            imp.Fields = fields
            root.Children = append(root.Children, imp)
        case astpkg.FuncDecl:
            fn := sch.ASTNode{Kind: "FuncDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            fn.Fields = map[string]interface{}{"name": n.Name}
            root.Children = append(root.Children, fn)
        case astpkg.PipelineDecl:
            pd := sch.ASTNode{Kind: "PipelineDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            pd.Fields = map[string]interface{}{"name": n.Name, "connectors": n.Connectors}
            // children: node calls
            for _, st := range n.Steps {
                call := sch.ASTNode{Kind: "NodeCall", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
                fields := map[string]interface{}{"name": st.Name}
                if len(st.Args) > 0 { fields["args"] = st.Args }
                if len(st.Workers) > 0 {
                    var ws []map[string]string
                    for _, w := range st.Workers { ws = append(ws, map[string]string{"name": w.Name, "kind": w.Kind}) }
                    fields["workers"] = ws
                }
                call.Fields = fields
                pd.Children = append(pd.Children, call)
            }
            if len(n.ErrorSteps) > 0 {
                errNode := sch.ASTNode{Kind: "ErrorPipeline", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
                errNode.Fields = map[string]interface{}{"connectors": n.ErrorConnectors}
                for _, st := range n.ErrorSteps {
                    call := sch.ASTNode{Kind: "NodeCall", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
                    fields := map[string]interface{}{"name": st.Name}
                    if len(st.Args) > 0 { fields["args"] = st.Args }
                    if len(st.Workers) > 0 {
                        var ws []map[string]string
                        for _, w := range st.Workers { ws = append(ws, map[string]string{"name": w.Name, "kind": w.Kind}) }
                        fields["workers"] = ws
                    }
                    call.Fields = fields
                    errNode.Children = append(errNode.Children, call)
                }
                pd.Children = append(pd.Children, errNode)
            }
            root.Children = append(root.Children, pd)
        case astpkg.EnumDecl:
            en := sch.ASTNode{Kind: "EnumDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            en.Fields = map[string]interface{}{"name": n.Name}
            root.Children = append(root.Children, en)
        case astpkg.StructDecl:
            st := sch.ASTNode{Kind: "StructDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            st.Fields = map[string]interface{}{"name": n.Name}
            root.Children = append(root.Children, st)
        case astpkg.Bad:
            bad := sch.ASTNode{Kind: "Bad", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            bad.Fields = map[string]interface{}{"token": n.Tok.Kind, "lexeme": n.Tok.Lexeme}
            root.Children = append(root.Children, bad)
        default:
            unk := sch.ASTNode{Kind: "UnknownDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
            root.Children = append(root.Children, unk)
        }
    }

    // directives (pragma) as children after package decl
    for _, dr := range file.Directives {
        dn := sch.ASTNode{Kind: "Directive", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}
        dn.Fields = map[string]interface{}{"name": dr.Name, "payload": dr.Payload}
        root.Children = append(root.Children, dn)
    }

    return sch.ASTV1{Schema: "ast.v1", Package: file.Package, File: filePath, Root: root}
}
