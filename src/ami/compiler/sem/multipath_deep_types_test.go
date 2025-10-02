package sem

import (
	"github.com/sam-caldwell/ami/src/ami/compiler/parser"
	"github.com/sam-caldwell/ami/src/ami/compiler/source"
	diag "github.com/sam-caldwell/ami/src/schemas/diag"
	"testing"
)

func hasCodeRec(ds []diag.Record, code string) bool {
	for _, d := range ds {
		if d.Code == code {
			return true
		}
	}
	return false
}

// Deep nested struct resolution: orderable when leaf is primitive.
func testMergeField_DeepNested_Orderable(t *testing.T) {
	// A: Event<Struct{a:Struct{b:Struct{c:string}}}>
	// Collect: merge.Sort("a.b.c") → orderable
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{a:Struct{b:Struct{c:string}}}>\"); Collect merge.Sort(\"a.b.c\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep1.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", ds)
	}
}

// Unknown deep path should report E_MERGE_SORT_FIELD_UNKNOWN.
func testMergeField_DeepNested_Unknown(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{a:Struct{b:Struct{c:string}}}>\"); Collect merge.Sort(\"a.b.d\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep2.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if !hasCodeRec(ds, "E_MERGE_SORT_FIELD_UNKNOWN") {
		t.Fatalf("expected E_MERGE_SORT_FIELD_UNKNOWN, got %+v", ds)
	}
}

// Non-primitive leaf types should be unorderable when resolved.
func testMergeField_NonPrimitive_Unorderable(t *testing.T) {
	// x is slice<int> so non-primitive → E_MERGE_SORT_FIELD_UNORDERABLE
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{x:slice<int>}>\"); Collect merge.Sort(\"x\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep3.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if !hasCodeRec(ds, "E_MERGE_SORT_FIELD_UNORDERABLE") {
		t.Fatalf("expected E_MERGE_SORT_FIELD_UNORDERABLE, got %+v", ds)
	}
}

// Multiple upstreams: resolving from any upstream with type info should suffice.
func testMergeField_MultipleUpstreams_ResolveFromAny(t *testing.T) {
	// A has the correct struct; B is untyped. A -> Collect, B -> Collect.
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{k:int}>\"); B; Collect merge.Sort(\"k\"); A -> Collect; B -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep4.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", ds)
	}
}

// Primitive payload with field reference should report E_MERGE_FIELD_ON_PRIMITIVE.
func testMergeField_PrimitivePayload_FieldReference_Error(t *testing.T) {
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<int>\"); Collect merge.Sort(\"k\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep5.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if !hasCodeRec(ds, "E_MERGE_FIELD_ON_PRIMITIVE") {
		t.Fatalf("expected E_MERGE_FIELD_ON_PRIMITIVE, got %+v", ds)
	}
}

// Optional/Union behavior is supported in the type system; deep cases below.
// Deep Optional+Union: orderable when union alts are primitives under Optional
func testMergeField_DeepOptionalUnion_Orderable(t *testing.T) {
	// Event<Struct{a:Optional<Struct{b:Union<int,string>}>}>, field a.b → Optional<Union<int,string>> orderable
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{a:Optional<Struct{b:Union<int,string>}>}>\"); Collect merge.Sort(\"a.b\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep_opt_union.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", ds)
	}
}

// Deep Union of Structs with differing fields: unknown when not all alts resolve
func testMergeField_DeepUnion_MissingField_Unknown(t *testing.T) {
	// Event<Struct{a:Union<Struct{b:int},Struct{c:int}>}>, field a.b → unknown
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{a:Union<Struct{b:int},Struct{c:int}>}>\"); Collect merge.Sort(\"a.b\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep_union_unknown.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if !hasCodeRec(ds, "E_MERGE_SORT_FIELD_UNKNOWN") {
		t.Fatalf("expected E_MERGE_SORT_FIELD_UNKNOWN, got %+v", ds)
	}
}

// Triple-nested deep path with Optional above Union leaf is orderable
func testMergeField_DeepNested_OptionalUnion_PrimitiveLeaf_Orderable(t *testing.T) {
	// Event<Struct{a:Optional<Struct{b:Struct{c:Union<int,string>}}}>>, field a.b.c → orderable
	code := "package app\n" +
		"pipeline P(){ A type(\"Event<Struct{a:Optional<Struct{b:Struct{c:Union<int,string>}}>}>\"); Collect merge.Sort(\"a.b.c\"); A -> Collect; egress }\n"
	f := (&source.FileSet{}).AddFile("deep_nested_opt_union.ami", code)
	af, _ := parser.New(f).ParseFile()
	ds := AnalyzeMergeFieldTypes(af)
	if len(ds) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", ds)
	}
}
