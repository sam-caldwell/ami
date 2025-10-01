package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure AMI stdlib stubs for package "math" are importable and core calls type-check.
func TestAMIStdlib_Math_Stubs_Resolve(t *testing.T) {
    ws := workspace.Workspace{}
    app := &source.FileSet{}
    src := "package app\nimport math\n" +
        "func F(){\n" +
        "  var x float64\n" +
        "  math.Abs(x)\n" +
        "  math.Ceil(x)\n" +
        "  math.Exp(x)\n" +
        "  math.Log2(x)\n" +
        "  math.Pow(x, x)\n" +
        "  math.Sqrt(x)\n" +
        "  math.Sin(x)\n" +
        "  math.Tanh(x)\n" +
        "  math.NaN()\n" +
        "  math.Inf(1)\n" +
        "  math.IsNaN(x)\n" +
        "  math.IsInf(x, 1)\n" +
        "  math.Signbit(x)\n" +
        "  math.Copysign(x, x)\n" +
        "  math.Nextafter(x, x)\n" +
        "}\n"
    app.AddFile("app.ami", src)
    pkgs := []Package{{Name: "app", Files: app}}
    _, diags := Compile(ws, pkgs, Options{Debug: false, EmitLLVMOnly: true})
    for _, d := range diags {
        if string(d.Level) == "error" {
            t.Fatalf("unexpected error diagnostic: %+v", d)
        }
    }
}
