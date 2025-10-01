package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure builtin AMI stdlib time package exposes Duration-based APIs and Ticker helpers.
func TestAMIStdlib_Time_Builtin_API_Surface(t *testing.T) {
    ws := workspace.Workspace{}
    // Use builtin stubs by not providing a time package here.
    appfs := &source.FileSet{}
    appSrc := "package app\nimport time\n" +
        "func F() {\n" +
        "  var d Duration = 1s\n" +
        "  time.Sleep(d)\n" +
        "  var t = time.Now()\n" +
        "  var _u = time.Add(t, d)\n" +
        "  var _d = time.Delta(t, t)\n" +
        "  var _s = time.Unix(t)\n" +
        "  var _n = time.UnixNano(t)\n" +
        "  var _s2 = t.Unix()\n" +
        "  var _n2 = t.UnixNano()\n" +
        "  var tk = time.NewTicker(d)\n" +
        "  time.TickerRegister(tk, F)\n" +
        "  time.TickerStart(tk)\n" +
        "  time.TickerStop(tk)\n" +
        "}\n"
    appfs.AddFile("app.ami", appSrc)
    pkgs := []Package{{Name: "app", Files: appfs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false, EmitLLVMOnly: true})
    for _, d := range diags {
        if string(d.Level) == "error" {
            t.Fatalf("unexpected error diagnostic: %+v", d)
        }
    }
}
