package llvm

import (
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
)

// Verify that linking multiple objects resolves cross-object symbols.
func TestLinkObjects_ResolvesSymbolsAcrossObjects(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found") }
    dir := filepath.Join("build", "test", "llvm_link_resolve")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    triple := DefaultTriple
    // Module A defines Foo
    aLL := 
        "; ModuleID = \"ami:a\"\n" +
        "target triple = \"" + triple + "\"\n\n" +
        "define i32 @Foo() {\nentry:\n  ret i32 42\n}\n"
    aPath := filepath.Join(dir, "a.ll")
    if err := os.WriteFile(aPath, []byte(aLL), 0o644); err != nil { t.Fatalf("write a.ll: %v", err) }
    aObj := filepath.Join(dir, "a.o")
    if err := CompileLLToObject(clang, aPath, aObj, triple); err != nil { t.Fatalf("compile a: %v", err) }
    // Module B declares Foo and calls it in main
    bLL := 
        "; ModuleID = \"ami:b\"\n" +
        "target triple = \"" + triple + "\"\n\n" +
        "declare i32 @Foo()\n\n" +
        "define i32 @main() {\nentry:\n  %0 = call i32 @Foo()\n  ret i32 0\n}\n"
    bPath := filepath.Join(dir, "b.ll")
    if err := os.WriteFile(bPath, []byte(bLL), 0o644); err != nil { t.Fatalf("write b.ll: %v", err) }
    bObj := filepath.Join(dir, "b.o")
    if err := CompileLLToObject(clang, bPath, bObj, triple); err != nil { t.Fatalf("compile b: %v", err) }
    // Link
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    if err := LinkObjects(clang, []string{aObj, bObj}, bin, triple); err != nil { t.Fatalf("link: %v", err) }
    // Execute
    if err := exec.Command(bin).Run(); err != nil { t.Fatalf("run: %v", err) }
}

