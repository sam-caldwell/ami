package json

import (
    "bytes"
    stdjson "encoding/json"
    "testing"
)

func TestMarshal_MapKeyOrdering_TopLevel(t *testing.T) {
    in := map[string]any{"b": 1, "a": 2, "c": map[string]any{"y":1, "x":2}}
    got, err := Marshal(in)
    if err != nil { t.Fatal(err) }
    want := []byte(`{"a":2,"b":1,"c":{"x":2,"y":1}}`)
    if !bytes.Equal(got, want) {
        t.Fatalf("got %s want %s", got, want)
    }
}

type inner struct {
    M map[string]int `json:"m"`
}
type outer struct {
    Name string   `json:"name"`
    L    []inner  `json:"l"`
}

func TestMarshal_NestedStructsAndArrays(t *testing.T) {
    o := outer{
        Name: "x",
        L: []inner{
            {M: map[string]int{"b": 2, "a": 1}},
            {M: map[string]int{"b": 4, "a": 3}},
        },
    }
    got, err := Marshal(o)
    if err != nil { t.Fatal(err) }
    // Objects (including structs) are canonicalized with lexicographically sorted keys.
    want := []byte(`{"l":[{"m":{"a":1,"b":2}},{"m":{"a":3,"b":4}}],"name":"x"}`)
    if !bytes.Equal(got, want) { t.Fatalf("got %s want %s", got, want) }
}

func TestMarshal_StableAcrossRuns(t *testing.T) {
    in := map[string]any{"z": 0, "y": []any{ map[string]int{"b":2, "a":1} }, "x": "s"}
    var first []byte
    for i := 0; i < 5; i++ {
        got, err := Marshal(in)
        if err != nil { t.Fatal(err) }
        if i == 0 { first = got; continue }
        if !bytes.Equal(got, first) { t.Fatal("non-deterministic marshal across runs") }
    }
}

func TestUnmarshalStrict_UnknownField_Error(t *testing.T) {
    type U struct { A int `json:"a"` }
    data := []byte(`{"a":1,"b":2}`)
    var u U
    if err := UnmarshalStrict(data, &u); err == nil { t.Fatal("expected unknown field error") }
}

func TestUnmarshal_NonStrict_IgnoresUnknown(t *testing.T) {
    type U struct { A int `json:"a"` }
    data := []byte(`{"a":1,"b":2}`)
    var u U
    if err := Unmarshal(data, &u); err != nil { t.Fatal(err) }
    if u.A != 1 { t.Fatalf("unexpected value: %+v", u) }
}

func TestMarshal_RespectsStructTags(t *testing.T) {
    type T struct { X int `json:"x"`; Y int `json:"-"` }
    b, err := Marshal(T{X:1, Y:2})
    if err != nil { t.Fatal(err) }
    if string(b) != `{"x":1}` { t.Fatalf("got %s", b) }
    // sanity: decoding with stdlib matches
    var m map[string]any
    if err := stdjson.Unmarshal(b, &m); err != nil { t.Fatal(err) }
    if len(m) != 1 || m["x"].(float64) != 1 { t.Fatalf("unexpected decode: %#v", m) }
}

func TestMarshal_ScalarsAndNull(t *testing.T) {
    // nil
    b, err := Marshal(nil)
    if err != nil || string(b) != "null" { t.Fatalf("nil got %s err %v", b, err) }
    // bool
    b, err = Marshal(true)
    if err != nil || string(b) != "true" { t.Fatalf("bool got %s err %v", b, err) }
    // string
    b, err = Marshal("hi")
    if err != nil || string(b) != "\"hi\"" { t.Fatalf("string got %s err %v", b, err) }
}

func TestMarshal_TopLevelArray_WithNestedObject(t *testing.T) {
    in := []any{ map[string]int{"b":2, "a":1}, "str", true, nil }
    got, err := Marshal(in)
    if err != nil { t.Fatal(err) }
    want := `[{"a":1,"b":2},"str",true,null]`
    if string(got) != want { t.Fatalf("got %s want %s", got, want) }
}

type customArrayMarshaler struct{}
func (customArrayMarshaler) MarshalJSON() ([]byte, error) { return []byte(`[{"z":1,"a":2}]`), nil }

func TestMarshal_CustomMarshaler_NestedFallback(t *testing.T) {
    b, err := Marshal(customArrayMarshaler{})
    if err != nil { t.Fatal(err) }
    if string(b) != `[{"a":2,"z":1}]` { t.Fatalf("got %s", b) }
}

// custom Marshaler to exercise the fallback path in canonical encoder
type customMarshaler struct{}
func (customMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{"z":1,"a":2}`), nil }

func TestMarshal_CustomMarshaler_FallbackSorted(t *testing.T) {
    b, err := Marshal(customMarshaler{})
    if err != nil { t.Fatal(err) }
    if string(b) != `{"a":2,"z":1}` { t.Fatalf("got %s", b) }
}

func TestUnmarshal_UseNumber(t *testing.T) {
    var m map[string]any
    if err := Unmarshal([]byte(`{"n": 1}`), &m); err != nil { t.Fatal(err) }
    if _, ok := m["n"].(stdjson.Number); !ok { t.Fatalf("number type not preserved: %T", m["n"]) }
}

func TestCanonicalize_And_Compact(t *testing.T) {
    raw := []byte(` {"b":1, "a": {"z":2, "y":3} } `)
    can, err := Canonicalize(raw)
    if err != nil { t.Fatal(err) }
    if string(can) != `{"a":{"y":3,"z":2},"b":1}` { t.Fatalf("canonical got %s", can) }
    cmp, err := Compact(can)
    if err != nil { t.Fatal(err) }
    if string(cmp) != string(can) { t.Fatalf("compact changed bytes: %s vs %s", cmp, can) }
}

func TestMarshalIndent(t *testing.T) {
    in := map[string]any{"b":1, "a":2}
    got, err := MarshalIndent(in, "", "  ")
    if err != nil { t.Fatal(err) }
    want := "{\n  \"a\": 2,\n  \"b\": 1\n}"
    if string(got) != want { t.Fatalf("got %q want %q", string(got), want) }
}

func TestEqualCanonical_And_Valid(t *testing.T) {
    a := []byte(`{"b":1, "a":2}`)
    b := []byte(`{"a":2,"b":1}`)
    eq, err := EqualCanonical(a, b)
    if err != nil { t.Fatal(err) }
    if !eq { t.Fatal("expected canonical equality") }
    if !Valid(a) || !Valid(b) { t.Fatal("expected valid JSON") }
    invalid := []byte(`{"a":}`)
    if Valid(invalid) { t.Fatal("expected invalid JSON") }
}

func TestMarshal_ErrorOnUnsupportedType(t *testing.T) {
    ch := make(chan int)
    if _, err := Marshal(ch); err == nil { t.Fatal("expected error for unsupported type") }
}

func TestCanonicalize_ErrorOnInvalidJSON(t *testing.T) {
    if _, err := Canonicalize([]byte(`{"a":}`)); err == nil { t.Fatal("expected error") }
}

func TestEqualCanonical_ErrorOnInvalid(t *testing.T) {
    if _, err := EqualCanonical([]byte(`}`), []byte(`{}`)); err == nil { t.Fatal("expected error") }
}

func TestCompact_ErrorOnInvalid(t *testing.T) {
    if _, err := Compact([]byte(`{"a":}`)); err == nil { t.Fatal("expected error") }
}
