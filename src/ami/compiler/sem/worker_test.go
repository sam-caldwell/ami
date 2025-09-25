package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestWorkers_Transform_ResolvesFunction(t *testing.T) {
	src := `package p
func f(ctx Context, ev Event<string>, st State) Event<string> {}
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_WORKER_UNDEFINED" {
			t.Fatalf("unexpected undefined worker: %+v", d)
		}
	}
}

func TestWorkers_Transform_ResolvesFactoryCall(t *testing.T) {
	src := `package p
func NewWorker(ctx Context, ev Event<string>, st State) Event<string> {}
pipeline P { Ingress(cfg).Transform(NewWorker(cfg)).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_WORKER_UNDEFINED" {
			t.Fatalf("unexpected undefined worker: %+v", d)
		}
	}
}

func TestWorkers_Transform_Undefined_Diagnostic(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Transform(doesNotExist()).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var seen bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_WORKER_UNDEFINED" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected E_WORKER_UNDEFINED; got %+v", res.Diagnostics)
	}
}

func TestWorkers_Transform_InvalidSignature_Diagnostic(t *testing.T) {
	src := `package p
func f(a int) int {}
pipeline P { Ingress(cfg).Transform(f).Egress(cfg) }`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	var seen bool
	for _, d := range res.Diagnostics {
		if d.Code == "E_WORKER_SIGNATURE" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected E_WORKER_SIGNATURE; got %+v", res.Diagnostics)
	}
}
