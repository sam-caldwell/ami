package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"testing"
)

func TestAnalyzeEdges_InvalidBackpressure(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Egress(in=edge.FIFO(minCapacity=1,maxCapacity=2,backpressure=hold,type=[]byte)) }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EDGE_BP_INVALID" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EDGE_BP_INVALID; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEdges_CapacityOrder(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Egress(in=edge.FIFO(minCapacity=3,maxCapacity=2,backpressure=block,type=T)) }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EDGE_CAP_ORDER" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EDGE_CAP_ORDER; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEdges_MinNegative(t *testing.T) {
src := `package p
pipeline P { Ingress(cfg).Egress(in=edge.LIFO(minCapacity=-1,maxCapacity=2,backpressure=dropOldest,type=T)) }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EDGE_MINCAP" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EDGE_MINCAP; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEdges_PipelineNameRequired(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Egress(in=edge.Pipeline(minCapacity=1,maxCapacity=2,backpressure=block,type=T)) }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == "E_EDGE_NAME_REQUIRED" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected E_EDGE_NAME_REQUIRED; diags=%v", res.Diagnostics)
	}
}

func TestAnalyzeEdges_Happy(t *testing.T) {
	src := `package p
pipeline P { Ingress(cfg).Egress(in=edge.FIFO(minCapacity=0,maxCapacity=0,backpressure=block,type=T)) }
`
	p := parser.New(src)
	f := p.ParseFile()
	res := AnalyzeFile(f)
	for _, d := range res.Diagnostics {
		if d.Code == "E_EDGE_MINCAP" || d.Code == "E_EDGE_CAP_ORDER" || d.Code == "E_EDGE_BP_INVALID" || d.Code == "E_EDGE_NAME_REQUIRED" {
			t.Fatalf("unexpected edge diagnostic: %v", d)
		}
	}
}
