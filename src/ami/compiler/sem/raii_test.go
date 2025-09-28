package sem

import (
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func TestRAII_DeferRelease_NoDouble(t *testing.T) {
    code := "package app\nfunc F(){ var x int; defer release(x) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    for _, d := range ds { if d.Code == "E_RAII_DOUBLE_RELEASE" { t.Fatalf("unexpected double-release: %+v", ds) } }
}

func TestRAII_DoubleRelease_ImmediateThenDefer(t *testing.T) {
    code := "package app\nfunc F(){ var y int; release(y); defer release(y) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_RAII_DOUBLE_RELEASE" { saw = true } }
    if !saw { t.Fatalf("expected E_RAII_DOUBLE_RELEASE, got: %+v", ds) }
}

func TestRAII_DoubleRelease_TwoDefers(t *testing.T) {
    code := "package app\nfunc F(){ var z int; defer release(z); defer release(z) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_RAII_DOUBLE_RELEASE" { saw = true } }
    if !saw { t.Fatalf("expected E_RAII_DOUBLE_RELEASE, got: %+v", ds) }
}

func TestRAII_MutateRelease_DeferMix(t *testing.T) {
    code := "package app\nfunc F(){ var a int; mutate(release(a)); defer mutate(release(a)) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    var saw bool
    for _, d := range ds { if d.Code == "E_RAII_DOUBLE_RELEASE" { saw = true } }
    if !saw { t.Fatalf("expected E_RAII_DOUBLE_RELEASE, got: %+v", ds) }
}

func TestRAII_UseAfterRelease_Immediate(t *testing.T) {
    // immediate release followed by use should emit E_RAII_USE_AFTER_RELEASE
    code := "package app\nfunc F(){ var a; release(a); g(a) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    saw := false
    for _, d := range ds { if d.Code == "E_RAII_USE_AFTER_RELEASE" { saw = true } }
    if !saw { t.Fatalf("expected E_RAII_USE_AFTER_RELEASE, got: %+v", ds) }
}

func TestRAII_Release_Unowned(t *testing.T) {
    // releasing undeclared variable should emit E_RAII_RELEASE_UNOWNED
    code := "package app\nfunc F(){ release(x) }\n"
    var fs source.FileSet
    f := fs.AddFile("f.ami", code)
    p := parser.New(f)
    af, errs := p.ParseFileCollect()
    if af == nil || len(errs) != 0 { t.Fatalf("parse: errs=%v", errs) }
    ds := AnalyzeRAII(af)
    saw := false
    for _, d := range ds { if d.Code == "E_RAII_RELEASE_UNOWNED" { saw = true } }
    if !saw { t.Fatalf("expected E_RAII_RELEASE_UNOWNED, got: %+v", ds) }
}
