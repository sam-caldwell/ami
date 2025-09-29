package main

import (
    "fmt"
    "os"
    "path/filepath"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
)

func main() {
    dir := filepath.Join("build", "test", "llvm_runtime_manual")
    if err := os.MkdirAll(dir, 0o755); err != nil { panic(err) }
    p, err := llvme.WriteRuntimeLL(dir, llvme.DefaultTriple, false)
    if err != nil { panic(err) }
    fmt.Print(p)
}

