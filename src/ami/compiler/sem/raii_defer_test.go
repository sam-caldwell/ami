package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/parser"
)

func TestRAII_Defer_MethodRelease_OK(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { defer r.Close() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_RAII_OWNED_NOT_RELEASED" { t.Fatalf("unexpected not-released: %v", d) } }
}

func TestRAII_Defer_FunctionRelease_OK(t *testing.T) {
    src := `package p
func release(o Owned<string>) {}
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { defer release(r) }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_RAII_OWNED_NOT_RELEASED" { t.Fatalf("unexpected not-released: %v", d) } }
}

func TestRAII_Defer_UseAllowedBeforeEnd_NoUseAfterError(t *testing.T) {
    src := `package p
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { defer r.Close(); r }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    for _, d := range res.Diagnostics { if d.Code == "E_RAII_USE_AFTER_RELEASE" { t.Fatalf("unexpected use-after-release: %v", d) } }
}

func TestRAII_Defer_DoubleRelease_Error(t *testing.T) {
    src := `package p
func release(o Owned<string>) {}
func f(ctx Context, ev Event<string>, r Owned<string>, st State) Event<string> { mutate(release(r)); defer r.Close() }`
    p := parser.New(src)
    f := p.ParseFile()
    res := AnalyzeFile(f)
    found := false
    for _, d := range res.Diagnostics { if d.Code == "E_RAII_DOUBLE_RELEASE" { found = true; break } }
    if !found { t.Fatalf("expected E_RAII_DOUBLE_RELEASE; diags=%v", res.Diagnostics) }
}
