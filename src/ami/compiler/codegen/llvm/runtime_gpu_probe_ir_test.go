package llvm

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Ensure runtime IR includes GPU probe initializer and mask, and darwin triples call metal available.
func TestRuntimeLL_GPUProbe_IncludesInit_Global_And_MetalOnDarwin(t *testing.T) {
    // Use a Darwin triple
    s := RuntimeLL(TripleFor("darwin", "arm64"), false)
    if !strings.Contains(s, "@ami_rt_gpu_mask = private global i64 0") {
        t.Fatalf("missing gpu mask global in IR:\n%s", s)
    }
    if !strings.Contains(s, "define void @ami_rt_gpu_probe_init()") {
        t.Fatalf("missing gpu probe init function in IR:\n%s", s)
    }
    if !strings.Contains(s, "define i1 @ami_rt_gpu_has(i64 %which)") {
        t.Fatalf("missing gpu has() accessor in IR:\n%s", s)
    }
    if !strings.Contains(s, "call i1 @ami_rt_metal_available()") {
        t.Fatalf("darwin triple did not include metal availability call in probe:\n%s", s)
    }
}

// Ensure non-darwin triples omit the metal availability call in probe function.
func TestRuntimeLL_GPUProbe_OmitsMetalCall_OnLinux(t *testing.T) {
    s := RuntimeLL(TripleFor("linux", "amd64"), false)
    if strings.Contains(s, "call i1 @ami_rt_metal_available()") {
        t.Fatalf("linux triple should not include metal availability call in probe:\n%s", s)
    }
}

// Ensure compile-time pruning via AMI_GPU_BACKENDS removes getenv gates for pruned backends.
func TestRuntimeLL_GPUProbe_PruneBackends_EnvFlag(t *testing.T) {
    // Keep a copy of existing value and restore
    old := os.Getenv("AMI_GPU_BACKENDS")
    defer os.Setenv("AMI_GPU_BACKENDS", old)
    if err := os.Setenv("AMI_GPU_BACKENDS", "metal"); err != nil {
        t.Fatalf("setenv: %v", err)
    }
    s := RuntimeLL(TripleFor("linux", "amd64"), false)
    if strings.Contains(s, "AMI_GPU_FORCE_OPENCL") || strings.Contains(s, "@.gpu.env.opencl") {
        t.Fatalf("opencl getenv gate present despite pruning:\n%s", s)
    }
    if strings.Contains(s, "AMI_GPU_FORCE_CUDA") || strings.Contains(s, "@.gpu.env.cuda") {
        t.Fatalf("cuda getenv gate present despite pruning:\n%s", s)
    }
}

// Verify entrypoint calls the probe init before spawning ingress.
func TestEntry_Writes_ProbeInit_Call(t *testing.T) {
    dir := filepath.Join("build", "test", "entry_gpu_probe")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    p, err := WriteIngressEntrypointLL(dir, DefaultTriple, []string{"pkg.pipe"})
    if err != nil { t.Fatalf("WriteIngressEntrypointLL: %v", err) }
    b, _ := os.ReadFile(p)
    s := string(b)
    if !strings.Contains(s, "declare void @ami_rt_gpu_probe_init()") {
        t.Fatalf("missing probe init declaration in entry IR:\n%s", s)
    }
    if !strings.Contains(s, "call void @ami_rt_gpu_probe_init()") {
        t.Fatalf("missing probe init call in entry IR:\n%s", s)
    }
}

