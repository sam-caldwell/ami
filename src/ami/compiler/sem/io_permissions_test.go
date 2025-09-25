package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestIOPermissions_TransformForbidden(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Transform(io=read).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	has := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_IO_PERMISSION" {
			has = true
			break
		}
	}
	if !has {
		t.Fatalf("expected E_IO_PERMISSION for Transform with io attribute; diags=%+v", res.Diagnostics)
	}
}

func TestIOPermissions_IngressEgressAllowed(t *testing.T) {
	src := `package p
pipeline P { Ingress(io=read).Egress(io=write) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_IO_PERMISSION" {
			t.Fatalf("unexpected E_IO_PERMISSION on ingress/egress: %+v", res.Diagnostics)
		}
	}
}
