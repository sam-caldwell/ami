package driver

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLower_GPU_Enumeration_EmitsExterns_AndCalls(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    src := "package app\nimport gpu\nfunc F(){ _ = gpu.CudaDevices(); _ = gpu.OpenCLPlatforms(); _ = gpu.OpenCLDevices() }\n"
    fs.AddFile("u.ami", src)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    s := string(b)
    decls := []string{
        "declare ptr @ami_rt_cuda_devices()",
        "declare ptr @ami_rt_opencl_platforms()",
        "declare ptr @ami_rt_opencl_devices()",
    }
    for _, d := range decls { if !strings.Contains(s, d) { t.Fatalf("missing extern: %s\n%s", d, s) } }
    calls := []string{
        "call ptr @ami_rt_cuda_devices()",
        "call ptr @ami_rt_opencl_platforms()",
        "call ptr @ami_rt_opencl_devices()",
    }
    for _, c := range calls { if !strings.Contains(s, c) { t.Fatalf("missing call: %s\n%s", c, s) } }
}

