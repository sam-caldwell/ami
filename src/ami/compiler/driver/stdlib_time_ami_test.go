package driver

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Provide AMI stdlib stubs for package "time" and ensure resolution works from app.
func TestAMIStdlib_Time_Stubs_Resolve(t *testing.T) {
    ws := workspace.Workspace{}
    // AMI stdlib stubs: package time
    timefs := &source.FileSet{}
    timeSrc := "package time\n" +
        "// AMI stdlib stubs (signatures only)\n" +
        "func Sleep(d int) {}\n" +
        "func Now() (Time) {}\n" +
        "func Add(t Time, d int) (Time) {}\n" +
        "func Delta(a Time, b Time) (int64) {}\n" +
        "func Unix(t Time) (int64) {}\n" +
        "func UnixNano(t Time) (int64) {}\n"
    timefs.AddFile("time.ami", timeSrc)

    // App using the time stubs
    appfs := &source.FileSet{}
    appSrc := "package app\nimport time\n" +
        "func F(){\n" +
        "  time.Sleep(1)\n" +
        "  var t = time.Now()\n" +
        "  var u = time.Add(t, 1)\n" +
        "  var _d = time.Delta(t, u)\n" +
        "  var _s = time.Unix(t)\n" +
        "  var _n = time.UnixNano(t)\n" +
        "}\n"
    appfs.AddFile("app.ami", appSrc)

    pkgs := []Package{{Name: "time", Files: timefs}, {Name: "app", Files: appfs}}
    _, diags := Compile(ws, pkgs, Options{Debug: false, EmitLLVMOnly: true})
    // Fail on any error-level diagnostics; warnings allowed.
    for _, d := range diags {
        if string(d.Level) == "error" {
            t.Fatalf("unexpected error diagnostic: %+v", d)
        }
    }
}
