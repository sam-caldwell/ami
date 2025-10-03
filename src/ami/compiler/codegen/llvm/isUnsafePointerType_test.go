package llvm

import "testing"

func Test_isUnsafePointerType(t *testing.T) {
    cases := map[string]bool{
        "ptr": true,
        " *int": true,
        "Pointer<string>": true,
        "pointer<Owned>": true,
        "int": false,
        "bool": false,
    }
    for in, want := range cases {
        if got := isUnsafePointerType(in); got != want {
            t.Fatalf("isUnsafePointerType(%q)=%v want %v", in, got, want)
        }
    }
}
