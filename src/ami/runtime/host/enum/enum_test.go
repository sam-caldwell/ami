package enum

import (
	"encoding/json"
	"testing"
)

func testEnum_Basics_StringOrdinalValidity(t *testing.T) {
	d := MustNewDescriptor("Color", []string{"Red", "Green", "Blue"})
	if !IsValid(d, 0) || !IsValid(d, 2) || IsValid(d, 3) {
		t.Fatal("IsValid failure")
	}
	if String(d, 1) != "Green" {
		t.Fatalf("String: %s", String(d, 1))
	}
	if Ordinal(2) != 2 {
		t.Fatal("Ordinal identity failed")
	}
	vals := Values(d)
	if len(vals) != 3 || vals[0] != 0 || vals[2] != 2 {
		t.Fatalf("Values: %+v", vals)
	}
	names := Names(d)
	if len(names) != 3 || names[0] != "Red" || names[2] != "Blue" {
		t.Fatalf("Names: %+v", names)
	}
}

func testEnum_Parse_RoundTrip_JSON_Text_GoString(t *testing.T) {
	d := MustNewDescriptor("Color", []string{"Red", "Green", "Blue"})
	// Parse/MustParse/FromOrdinal
	if v, err := Parse(d, "Green"); err != nil || v != 1 {
		t.Fatalf("Parse: %v %d", err, v)
	}
	if v := MustParse(d, "Red"); v != 0 {
		t.Fatalf("MustParse: %d", v)
	}
	if _, err := Parse(d, "green"); err == nil {
		t.Fatalf("expected E_ENUM_PARSE on case mismatch")
	}
	if _, err := FromOrdinal(d, 3); err == nil {
		t.Fatalf("expected E_ENUM_ORDINAL on out of range")
	}
	// JSON/Text marshal/unmarshal
	vv := Value{D: d, V: 2}
	b, err := json.Marshal(vv)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	if string(b) != `"Blue"` {
		t.Fatalf("json: %s", string(b))
	}
	var vv2 Value
	vv2.D = d
	if err := json.Unmarshal([]byte(`"Red"`), &vv2); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if vv2.V != 0 {
		t.Fatalf("vv2: %d", vv2.V)
	}
	if vv.GoString() != "Color(Blue)" {
		t.Fatalf("GoString: %s", vv.GoString())
	}
	// Text
	tb, err := vv.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(tb) != "Blue" {
		t.Fatalf("text: %s", string(tb))
	}
	var vv3 Value
	vv3.D = d
	if err := vv3.UnmarshalText([]byte("Green")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if vv3.V != 1 {
		t.Fatalf("vv3: %d", vv3.V)
	}
}

func testEnum_NewDescriptor_Duplicate_And_Empty(t *testing.T) {
	if _, err := NewDescriptor("X", []string{"A", "A"}); err == nil {
		t.Fatalf("expected duplicate error")
	}
	if _, err := NewDescriptor("X", []string{""}); err == nil {
		t.Fatalf("expected empty name error")
	}
}
