package llvm

import "testing"

func TestNewModuleEmitter_DefaultTriple_AndSetTargetTriple(t *testing.T) {
    e := NewModuleEmitter("p", "u")
    if e.triple != DefaultTriple { t.Fatalf("default triple mismatch: %s", e.triple) }
    e.SetTargetTriple("x86_64-unknown-linux-gnu")
    if e.triple != "x86_64-unknown-linux-gnu" { t.Fatalf("set triple failed: %s", e.triple) }
}

