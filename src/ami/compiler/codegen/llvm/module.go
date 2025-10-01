package llvm

import (
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ir"
)

// ModuleEmitter builds a textual LLVM module for a single package/unit.
type ModuleEmitter struct {
    pkg    string
    unit   string
    triple string
    funcs  []string
    types  map[string]struct{}
    externs []string
}

// DefaultTriple is the baseline target triple for darwin/arm64.
const DefaultTriple = "arm64-apple-macosx"

// NewModuleEmitter constructs a new emitter for the given package/unit.
func NewModuleEmitter(pkg, unit string) *ModuleEmitter {
    return &ModuleEmitter{pkg: pkg, unit: unit, triple: DefaultTriple, types: map[string]struct{}{}}
}

// SetTargetTriple sets the LLVM target triple.
func (e *ModuleEmitter) SetTargetTriple(triple string) { if triple != "" { e.triple = triple } }

// RequireExtern adds a declaration to the module if not already present.
func (e *ModuleEmitter) RequireExtern(decl string) {
    for _, d := range e.externs { if d == decl { return } }
    e.externs = append(e.externs, decl)
}

// AddFunction lowers a single IR function and appends it to the module.
func (e *ModuleEmitter) AddFunction(fn ir.Function) error {
    s, err := lowerFunction(fn)
    if err != nil { return err }
    e.funcs = append(e.funcs, s)
    return nil
}

// Build assembles the full textual LLVM module as a string.
func (e *ModuleEmitter) Build() string {
    var b strings.Builder
    // Header (deterministic)
    b.WriteString("; ModuleID = \"ami:")
    b.WriteString(e.pkg)
    if e.unit != "" { b.WriteString("/"); b.WriteString(e.unit) }
    b.WriteString("\"\n")
    b.WriteString("target triple = \"")
    b.WriteString(e.triple)
    b.WriteString("\"\n\n")
    // Minimal runtime externs (deterministic order)
    ex := e.externs
    if len(ex) > 0 {
        // Ensure deterministic extern ordering regardless of discovery order.
        sort.Strings(ex)
        for _, d := range ex { b.WriteString(d); b.WriteString("\n") }
        b.WriteString("\n")
    }
    // Opaque/alias types (sorted for determinism); currently only comments for scaffolding.
    if len(e.types) > 0 {
        var keys []string
        for k := range e.types { keys = append(keys, k) }
        sort.Strings(keys)
        for _, k := range keys {
            b.WriteString("; type ")
            b.WriteString(k)
            b.WriteString(" is opaque (runtime handle)\n")
        }
        b.WriteString("\n")
    }
    // Functions
    for _, f := range e.funcs {
        b.WriteString(f)
        if !strings.HasSuffix(f, "\n") { b.WriteString("\n") }
        b.WriteString("\n")
    }
    return b.String()
}
