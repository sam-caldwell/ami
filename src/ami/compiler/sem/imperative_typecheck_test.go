package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestImperative_AssignTypeMismatch_Error(t *testing.T) {
	src := `package p
func f(a int, b string) {
    *a = b
}`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_ASSIGN_TYPE_MISMATCH" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_ASSIGN_TYPE_MISMATCH; diags=%v", res.Diagnostics)
	}
}

func TestImperative_AssignPointerDeref_Mismatch_Error(t *testing.T) {
	t.Skip("no raw pointer semantics in AMI 2.3.2")
}

func TestImperative_AssignPointerDeref_OK(t *testing.T) {
	t.Skip("no raw pointer semantics in AMI 2.3.2")
}

func TestImperative_AssignAddressOf_OK(t *testing.T) {
	t.Skip("no address-of in AMI 2.3.2")
}
