package driver

import "testing"

func TestArtifacts_StructFields(t *testing.T) {
    a := Artifacts{}
    if a.IR != nil { t.Fatalf("zero value should be nil slice: %+v", a) }
    a.IR = []string{"build/debug/ir/main.json"}
    if len(a.IR) != 1 || a.IR[0] == "" { t.Fatalf("unexpected IR: %+v", a.IR) }
}

