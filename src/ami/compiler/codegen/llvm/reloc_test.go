package llvm

import (
    "debug/elf"
    "debug/macho"
    "os"
    "path/filepath"
    "runtime"
    "testing"
)

// Test that compiling an LLVM IR module which calls an external function
// produces an object file with relocation information or at least an undefined symbol.
func TestCompileLLToObject_EmitsRelocations(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found") }

    dir := filepath.Join("build", "test", "llvm_reloc")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    triple := TripleFor(runtime.GOOS, runtime.GOARCH)
    // Construct a small module that references an external symbol (puts)
    ll := 
        "; ModuleID = \"ami:test-reloc\"\n" +
        "target triple = \"" + triple + "\"\n\n" +
        "declare i32 @puts(ptr)\n\n" +
        "@.str = private unnamed_addr constant [6 x i8] c\"hello\\00\", align 1\n\n" +
        "define i32 @main() {\n" +
        "entry:\n  %0 = getelementptr inbounds [6 x i8], ptr @.str, i32 0, i32 0\n" +
        "  %1 = call i32 @puts(ptr %0)\n" +
        "  ret i32 0\n}\n"

    llPath := filepath.Join(dir, "t.ll")
    if err := os.WriteFile(llPath, []byte(ll), 0o644); err != nil { t.Fatalf("write ll: %v", err) }
    obj := filepath.Join(dir, "t.o")
    if err := CompileLLToObject(clang, llPath, obj, triple); err != nil { t.Fatalf("compile: %v", err) }

    // macOS: Mach-O relocations
    if runtime.GOOS == "darwin" {
        f, err := macho.Open(obj)
        if err != nil { t.Fatalf("macho open: %v", err) }
        defer f.Close()
        hasRel := false
        for _, s := range f.Sections {
            if len(s.Relocs) > 0 { hasRel = true; break }
        }
        if !hasRel && f.Symtab != nil {
            for _, sym := range f.Symtab.Syms {
                if sym.Name == "_puts" && sym.Sect == 0 { hasRel = true; break }
            }
        }
        if !hasRel { t.Fatalf("expected relocations or undefined externs in Mach-O") }
        return
    }
    // Linux/others: ELF relocations
    f, err := elf.Open(obj)
    if err != nil { t.Skip("non-ELF target; skipping relocation validation") }
    defer f.Close()
    hasRel := false
    for _, s := range f.Sections {
        if s.Type == elf.SHT_RELA || s.Type == elf.SHT_REL {
            if s.Size > 0 { hasRel = true; break }
        }
    }
    if !hasRel { t.Fatalf("expected relocation sections in ELF object") }
}
