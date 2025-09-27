package sem

import (
    "encoding/json"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
)

func diagJSON(diags any) string { b, _ := json.Marshal(diags); return string(b) }

func TestMemSafety_BanAmpersand(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "x = &y"}
    ds := AnalyzeMemorySafety(f)
    if len(ds) != 1 { t.Fatalf("want 1 diag, got %d: %s", len(ds), diagJSON(ds)) }
    if ds[0].Code != "E_PTR_UNSUPPORTED_SYNTAX" { t.Fatalf("code: %s", ds[0].Code) }
}

func TestMemSafety_UnaryStarOnlyOnLHS(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "*x = 1"}
    ds := AnalyzeMemorySafety(f)
    if len(ds) != 0 { t.Fatalf("want 0 diags, got %d: %s", len(ds), diagJSON(ds)) }
}

func TestMemSafety_UnaryStarElsewhere_IsError(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "*f(1)"}
    ds := AnalyzeMemorySafety(f)
    if len(ds) != 1 { t.Fatalf("want 1 diag, got %d: %s", len(ds), diagJSON(ds)) }
    if ds[0].Code != "E_MUT_BLOCK_UNSUPPORTED" { t.Fatalf("code: %s", ds[0].Code) }
}

func TestMemSafety_BinaryMultiplyIsOK(t *testing.T) {
    f := &source.File{Name: "t.ami", Content: "a*b"}
    ds := AnalyzeMemorySafety(f)
    if len(ds) != 0 { t.Fatalf("want 0 diags, got %d: %s", len(ds), diagJSON(ds)) }
}

