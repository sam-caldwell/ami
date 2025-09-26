package astjson

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    sch "github.com/sam-caldwell/ami/src/schemas"
)

// ToSchemaAST converts an internal AST File into schemas.ASTV1 with
// a richer structural tree suitable for debug artifacts.
func ToSchemaAST(file *astpkg.File, filePath string) sch.ASTV1 {
    root := sch.ASTNode{Kind: "File", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}}

    if pkg := buildPackageNode(file.Package); pkg.Kind != "" {
        root.Children = append(root.Children, pkg)
    }

    for _, d := range file.Decls {
        switch n := d.(type) {
        case astpkg.ImportDecl:
            root.Children = append(root.Children, buildImportNode(n))
        case astpkg.FuncDecl:
            root.Children = append(root.Children, buildFuncNode(n))
        case astpkg.PipelineDecl:
            root.Children = append(root.Children, buildPipelineNode(n))
        case astpkg.EnumDecl:
            root.Children = append(root.Children, buildEnumNode(n))
        case astpkg.StructDecl:
            root.Children = append(root.Children, buildStructNode(n))
        case astpkg.Bad:
            root.Children = append(root.Children, buildBadNode(n))
        default:
            root.Children = append(root.Children, sch.ASTNode{Kind: "UnknownDecl", Pos: sch.Position{Line: 1, Column: 1, Offset: 0}})
        }
    }

    for _, dr := range file.Directives {
        root.Children = append(root.Children, buildDirectiveNode(dr))
    }

    return sch.ASTV1{Schema: "ast.v1", Package: file.Package, Version: file.Version, File: filePath, Root: root}
}

