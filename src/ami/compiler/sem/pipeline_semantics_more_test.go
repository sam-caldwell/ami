package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestAnalyze_Pipeline_EgressInMiddle_Diagnostic(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Egress(cfg).Transform(f).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var pos, dup bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_EGRESS_POSITION" {
			pos = true
		}
		if d.Code == "E_DUP_EGRESS" {
			dup = true
		}
	}
	if !(pos && dup) {
		t.Fatalf("expected E_EGRESS_POSITION and E_DUP_EGRESS; got %+v", res.Diagnostics)
	}
}

func TestAnalyze_Pipeline_IngressInMiddle_Diagnostic(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Transform(f).Ingress(cfg).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var pos, dup bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_INGRESS_POSITION" {
			pos = true
		}
		if d.Code == "E_DUP_INGRESS" {
			dup = true
		}
	}
	if !(pos && dup) {
		t.Fatalf("expected E_INGRESS_POSITION and E_DUP_INGRESS; got %+v", res.Diagnostics)
	}
}

func TestAnalyze_Pipeline_ErrorPath_MustEndWithEgress(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) error { Collect() } }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var seen bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_ERRPIPE_END_EGRESS" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected E_ERRPIPE_END_EGRESS; got %+v", res.Diagnostics)
	}
}

func TestAnalyze_Pipeline_ErrorPath_CannotStartWithIngress(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) error { Ingress(cfg).Egress(cfg) } }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var seen bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_ERRPIPE_START_INVALID" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected E_ERRPIPE_START_INVALID; got %+v", res.Diagnostics)
	}
}
