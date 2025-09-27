package parser

import (
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestParser_ParseFile_PackageDecl(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "package app"}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if file.PackageName != "app" { t.Fatalf("want package app, got %q", file.PackageName) }
}

func TestParser_ParseFile_ErrorsOnMissingKeyword(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "pkg app"}
    p := New(f)
    if _, err := p.ParseFile(); err == nil {
        t.Fatalf("expected error for missing 'package' keyword")
    }
}

func TestParser_Imports_And_Func(t *testing.T) {
    src := "package app\nimport alpha\nimport \"beta\"\nfunc main() {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if file.PackageName != "app" { t.Fatalf("pkg: %q", file.PackageName) }
    if len(file.Decls) != 3 { t.Fatalf("want 3 decls (2 imports, 1 func), got %d", len(file.Decls)) }
}

func TestParser_Func_Params_Results_And_Body(t *testing.T) {
    // typed params/returns and statements: var, call, return
    src := "package app\nfunc F(a T, b U) (R1,R2) { var x T; Alpha(); return a,b }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.Results) != 2 { t.Fatalf("want 2 results, got %d", len(fn.Results)) }
    // Find the return statement and validate tuple arity
    if fn.Body == nil { t.Fatalf("no body") }
    var saw bool
    for _, st := range fn.Body.Stmts {
        if rs, ok := st.(*ast.ReturnStmt); ok {
            if len(rs.Results) != 2 { t.Fatalf("want 2 return exprs, got %d", len(rs.Results)) }
            saw = true
        }
    }
    if !saw { t.Fatalf("no return found") }
}

func TestParser_Pipeline_And_ErrorBlock(t *testing.T) {
    // Pipeline with inner error block; and a top-level error block
    src := "package app\n// leading\npipeline P() { error {} }\nerror {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 decls, got %d", len(file.Decls)) }
}

func TestParser_Tolerant_Recovery_MultipleErrors(t *testing.T) {
    // Two malformed lines: bad import path and bad func header; expect both errors collected
    src := "package app\nimport 123\nfunc ( {\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    _, errs := p.ParseFileCollect()
    if len(errs) < 2 {
        t.Fatalf("expected at least 2 errors, got %d: %+v", len(errs), errs)
    }
}

func TestParser_Pipeline_Steps_With_Args(t *testing.T) {
    src := "package app\npipeline P() {\n  // step 1\n  Alpha() attr1, attr2(\"p\")\n  Beta(\"x\", y) ;\n  A -> B;\n}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
}

func TestParser_Pipeline_Attr_DottedNames(t *testing.T) {
    src := "package app\npipeline P() { Alpha() edge.MultiPath(merge.Sort(\"ts\"), merge.Stable()) }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    // Verify attr args reduce to dotted names
    if len(file.Decls) != 1 { t.Fatalf("decls: %d", len(file.Decls)) }
    pd, ok := file.Decls[0].(*ast.PipelineDecl)
    if !ok { t.Fatalf("decl0: %T", file.Decls[0]) }
    var found bool
    for _, s := range pd.Stmts {
        if st, ok := s.(*ast.StepStmt); ok {
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" {
                    if len(at.Args) != 2 { t.Fatalf("args len: %d", len(at.Args)) }
                    if at.Args[0].Text != "merge.Sort(…)" || at.Args[1].Text != "merge.Stable()" {
                        t.Fatalf("attr args: %+v", at.Args)
                    }
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("edge.MultiPath not found") }
}

func TestParser_Func_Assign_Binary_Defer(t *testing.T) {
    src := "package app\nfunc G(){ x = 1+2*3; defer Alpha(); }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("parse: %v", err) }
}

func TestParser_Import_WithVersionConstraint(t *testing.T) {
    src := "package app\nimport alpha >= v1.2.3\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    imp, ok := file.Decls[0].(*ast.ImportDecl)
    if !ok { t.Fatalf("first decl not ImportDecl: %T", file.Decls[0]) }
    if imp.Path != "alpha" { t.Fatalf("path: %q", imp.Path) }
    if imp.Constraint != ">= v1.2.3" { t.Fatalf("constraint: %q", imp.Constraint) }
}

func TestParser_Import_BlockForm_WithAndWithoutConstraints(t *testing.T) {
    src := "package app\nimport (\n alpha >= v1.2.3\n \"beta\"\n)\n"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 2 { t.Fatalf("want 2 imports, got %d", len(file.Decls)) }
    imp0, ok := file.Decls[0].(*ast.ImportDecl)
    if !ok { t.Fatalf("decl0 not ImportDecl: %T", file.Decls[0]) }
    if imp0.Path != "alpha" || imp0.Constraint != ">= v1.2.3" { t.Fatalf("decl0 unexpected: %+v", imp0) }
    imp1, ok := file.Decls[1].(*ast.ImportDecl)
    if !ok { t.Fatalf("decl1 not ImportDecl: %T", file.Decls[1]) }
    if imp1.Path != "beta" || imp1.Constraint != "" { t.Fatalf("decl1 unexpected: %+v", imp1) }
}

func TestParser_Func_TypeParams_NoConstraint(t *testing.T) {
    src := "package app\nfunc F<T>(a T) {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.TypeParams) != 1 || fn.TypeParams[0].Name != "T" || fn.TypeParams[0].Constraint != "" {
        t.Fatalf("unexpected type params: %+v", fn.TypeParams)
    }
}

func TestParser_Func_TypeParams_WithConstraint_AndMultiple(t *testing.T) {
    src := "package app\nfunc F<T any, U>(a T, b U) {}"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl is %T", file.Decls[0]) }
    if len(fn.TypeParams) != 2 { t.Fatalf("want 2 type params, got %d", len(fn.TypeParams)) }
    if fn.TypeParams[0].Name != "T" || fn.TypeParams[0].Constraint != "any" { t.Fatalf("tp0: %+v", fn.TypeParams[0]) }
    if fn.TypeParams[1].Name != "U" || fn.TypeParams[1].Constraint != "" { t.Fatalf("tp1: %+v", fn.TypeParams[1]) }
}

func TestParser_ContainerLiterals_Slice_Set_Map(t *testing.T) {
    src := "package app\nfunc F(){ x = slice<T>{1,2}; y = set<T>{\"a\"}; z = map<K,V>{\"k\": 3} }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    if len(file.Decls) != 1 { t.Fatalf("want 1 decl, got %d", len(file.Decls)) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok { t.Fatalf("decl type: %T", file.Decls[0]) }
    if fn.Body == nil || len(fn.Body.Stmts) < 3 { t.Fatalf("expected 3+ statements") }
    // We just ensure parsing didn’t error; deeper assertions would require walking Exprs.
}

func TestParser_Fixture_EBNFSample(t *testing.T) {
    b, err := os.ReadFile("testdata/ebnf_sample.ami")
    if err != nil { t.Fatalf("read: %v", err) }
    f := &source.File{Name: "ebnf_sample.ami", Content: string(b)}
    p := New(f)
    if _, err := p.ParseFile(); err != nil { t.Fatalf("parse: %v", err) }
}

func TestParser_Positions_Expressions(t *testing.T) {
    src := "package app\nfunc F(){ var x T; x = slice<T>{1}; a.b; Alpha() }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func decl or body") }
    if len(fn.Body.Stmts) < 4 { t.Fatalf("expected 4 stmts, got %d", len(fn.Body.Stmts)) }
    // var x T
    if vd, ok := fn.Body.Stmts[0].(*ast.VarDecl); ok {
        if vd.Pos.Line == 0 || vd.NamePos.Line == 0 { t.Fatalf("var positions missing: %+v", vd) }
    } else { t.Fatalf("stmt0 not VarDecl: %T", fn.Body.Stmts[0]) }
    // x = slice<T>{1}
    if as, ok := fn.Body.Stmts[1].(*ast.AssignStmt); ok {
        if as.Pos.Line == 0 || as.NamePos.Line == 0 { t.Fatalf("assign positions missing: %+v", as) }
        if sl, ok := as.Value.(*ast.SliceLit); ok {
            if sl.Pos.Line == 0 || sl.LBrace.Line == 0 || sl.RBrace.Line == 0 { t.Fatalf("slice lit positions missing: %+v", sl) }
        } else { t.Fatalf("assign value not SliceLit: %T", as.Value) }
    } else { t.Fatalf("stmt1 not AssignStmt: %T", fn.Body.Stmts[1]) }
    // a.b
    if es, ok := fn.Body.Stmts[2].(*ast.ExprStmt); ok {
        if sel, ok := es.X.(*ast.SelectorExpr); ok {
            if sel.Pos.Line == 0 || sel.SelPos.Line == 0 { t.Fatalf("selector positions missing: %+v", sel) }
        } else { t.Fatalf("expr not SelectorExpr: %T", es.X) }
    } else { t.Fatalf("stmt2 not ExprStmt: %T", fn.Body.Stmts[2]) }
    // Alpha()
    if es, ok := fn.Body.Stmts[3].(*ast.ExprStmt); ok {
        if ce, ok := es.X.(*ast.CallExpr); ok {
            if ce.Pos.Line == 0 || ce.LParen.Line == 0 || ce.RParen.Line == 0 { t.Fatalf("call positions missing: %+v", ce) }
        } else { t.Fatalf("expr not CallExpr: %T", es.X) }
    } else { t.Fatalf("stmt3 not ExprStmt: %T", fn.Body.Stmts[3]) }
}

func TestParser_Call_DottedName(t *testing.T) {
    src := "package app\nfunc F(){ a.b() }"
    f := &source.File{Name: "t.ami", Content: src}
    p := New(f)
    file, err := p.ParseFile()
    if err != nil { t.Fatalf("parse: %v", err) }
    fn, ok := file.Decls[0].(*ast.FuncDecl)
    if !ok || fn.Body == nil { t.Fatalf("no func decl/body") }
    es, ok := fn.Body.Stmts[0].(*ast.ExprStmt)
    if !ok { t.Fatalf("stmt0 not ExprStmt: %T", fn.Body.Stmts[0]) }
    ce, ok := es.X.(*ast.CallExpr)
    if !ok { t.Fatalf("expr not CallExpr: %T", es.X) }
    if ce.Name != "a.b" { t.Fatalf("call name: %q", ce.Name) }
}
