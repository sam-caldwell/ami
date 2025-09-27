package workspace

import (
    "path/filepath"
    "testing"
)

func TestWorkspace_Validate_Defaults_OK(t *testing.T) {
    w := DefaultWorkspace()
    if errs := w.Validate(); len(errs) != 0 {
        t.Fatalf("unexpected errors: %v", errs)
    }
}

func TestWorkspace_Validate_TargetAbsolute_Err(t *testing.T) {
    w := DefaultWorkspace()
    w.Toolchain.Compiler.Target = filepath.Join(string([]rune{'/'}), "abs")
    if errs := w.Validate(); len(errs) == 0 {
        t.Fatalf("expected error for absolute target")
    }
}

func TestWorkspace_Validate_BadConcurrency_Err(t *testing.T) {
    w := DefaultWorkspace()
    w.Toolchain.Compiler.Concurrency = "0"
    if errs := w.Validate(); len(errs) == 0 {
        t.Fatalf("expected error for concurrency=0")
    }
}

func TestWorkspace_Validate_EnvPattern_Err(t *testing.T) {
    w := DefaultWorkspace()
    w.Toolchain.Compiler.Env = []string{"linux/amd64", "bad"}
    errs := w.Validate()
    found := false
    for _, e := range errs { if e == "toolchain.compiler.env[1] must be os/arch" { found = true } }
    if !found { t.Fatalf("expected env[1] error; got: %v", errs) }
}

