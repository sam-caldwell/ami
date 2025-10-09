package parser

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_CrossPkg_Types_Param_Result_Var(t *testing.T) {
    var fs source.FileSet
    code := "package app\nfunc F(a bufio.Reader) (bufio.Writer) { var x bufio.Scanner; return x }\n"
    fs.AddFile("u.ami", code)
    p := New(fs.Files[0])
    f, err := p.ParseFile()
    if err != nil || f == nil { t.Fatalf("parse: %v", err) }
    if len(f.Decls) == 0 { t.Fatalf("no decls") }
    fn := f.Decls[0].(*ast.FuncDecl)
    if fn.Params[0].Type != "bufio.Reader" { t.Fatalf("param type: %q", fn.Params[0].Type) }
    if fn.Results[0].Type != "bufio.Writer" { t.Fatalf("result type: %q", fn.Results[0].Type) }
    if vd, ok := fn.Body.Stmts[0].(*ast.VarDecl); ok {
        if vd.Type != "bufio.Scanner" { t.Fatalf("var type: %q", vd.Type) }
    } else { t.Fatalf("stmt0 not VarDecl: %T", fn.Body.Stmts[0]) }
}

func TestParser_CrossPkg_Types_In_Literals(t *testing.T) {
    var fs source.FileSet
    code := "package app\nfunc F(){ slice<bufio.Reader>{}; map<bufio.Reader, bufio.Writer>{} }\n"
    fs.AddFile("u2.ami", code)
    p := New(fs.Files[0])
    f, err := p.ParseFile()
    if err != nil || f == nil { t.Fatalf("parse: %v", err) }
    if len(f.Decls) == 0 { t.Fatalf("no decls") }
    fn := f.Decls[0].(*ast.FuncDecl)
    if len(fn.Body.Stmts) < 2 { t.Fatalf("expected 2 stmts, got %d", len(fn.Body.Stmts)) }
    if es, ok := fn.Body.Stmts[0].(*ast.ExprStmt); ok {
        if sl, ok2 := es.X.(*ast.SliceLit); ok2 {
            if sl.TypeName != "bufio.Reader" { t.Fatalf("slice elem type: %q", sl.TypeName) }
        } else { t.Fatalf("stmt0 expr not SliceLit: %T", es.X) }
    } else { t.Fatalf("stmt0 not ExprStmt: %T", fn.Body.Stmts[0]) }
    if es, ok := fn.Body.Stmts[1].(*ast.ExprStmt); ok {
        if ml, ok2 := es.X.(*ast.MapLit); ok2 {
            if ml.KeyType != "bufio.Reader" || ml.ValType != "bufio.Writer" {
                t.Fatalf("map types: %q -> %q", ml.KeyType, ml.ValType)
            }
        } else { t.Fatalf("stmt1 expr not MapLit: %T", es.X) }
    } else { t.Fatalf("stmt1 not ExprStmt: %T", fn.Body.Stmts[1]) }
}
