package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestEventTypeFlow_Mismatch_Error(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {}
func g(ev Event<int>) (Event<int>, error) {}
pipeline P { Ingress(cfg).Transform(f).Transform(g).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EVENT_TYPE_FLOW" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EVENT_TYPE_FLOW; diags=%v", res.Diagnostics)
	}
}

func TestEventTypeFlow_Match_OK(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {}
func g(ev Event<string>) (Event<string>, error) {}
pipeline P { Ingress(cfg).Transform(f).Transform(g).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_EVENT_TYPE_FLOW" {
			t.Fatalf("unexpected E_EVENT_TYPE_FLOW: %v", d)
		}
	}
}

func TestEventParam_Immutable_Assign_Error(t *testing.T) {
	src := `package p
func f(ev Event<string>) (Event<string>, error) {
    ev = ev
}
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EVENT_PARAM_ASSIGN" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EVENT_PARAM_ASSIGN; diags=%v", res.Diagnostics)
	}
}
