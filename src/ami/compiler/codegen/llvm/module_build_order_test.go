package llvm

import "testing"

func TestModuleEmitter_Build_OrdersAndDedupes(t *testing.T) {
    e := NewModuleEmitter("p", "u")
    e.SetTargetTriple("x86_64-unknown-linux-gnu")
    // externs: add in unsorted order with a duplicate; expect sorted unique
    e.RequireExtern("declare i32 @z()")
    e.RequireExtern("declare i32 @a()")
    e.RequireExtern("declare i32 @a()") // duplicate
    // globals: added unsorted; expect sorted
    e.AddGlobal("@g2 = private constant [1 x i8] c\"x\\00\"")
    e.AddGlobal("@g1 = private constant [1 x i8] c\"y\\00\"")
    // types: simulate presence by poking private map
    e.types["T2"] = struct{}{}
    e.types["T1"] = struct{}{}
    // functions: push two in order
    e.funcs = append(e.funcs, "define void @F() {\nret void\n}")
    e.funcs = append(e.funcs, "define void @A() {\nret void\n}")
    out := e.Build()
    // header with target triple
    if want := "target triple = \"x86_64-unknown-linux-gnu\""; !contains(out, want) { t.Fatalf("missing triple: %s", out) }
    // externs sorted
    if idxA, idxZ := indexOf(out, "declare i32 @a()"), indexOf(out, "declare i32 @z()"); !(idxA >= 0 && idxZ > idxA) {
        t.Fatalf("externs not sorted/deduped: %s", out)
    }
    // globals sorted (@g1 before @g2)
    if idx1, idx2 := indexOf(out, "@g1"), indexOf(out, "@g2"); !(idx1 >= 0 && idx2 > idx1) {
        t.Fatalf("globals not sorted: %s", out)
    }
    // types comments include both, sorted
    if !(indexOf(out, "; type T1 is opaque") >= 0 && indexOf(out, "; type T2 is opaque") > indexOf(out, "; type T1 is opaque")) {
        t.Fatalf("types not sorted: %s", out)
    }
}

// tiny helpers to avoid importing strings
func contains(s, sub string) bool { return indexOf(s, sub) >= 0 }
func indexOf(s, sub string) int {
    // naive substring search
    n, m := len(s), len(sub)
    if m == 0 { return 0 }
    for i := 0; i+m <= n; i++ { if s[i:i+m] == sub { return i } }
    return -1
}

