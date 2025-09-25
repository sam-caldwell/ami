package json

import (
    stdjson "encoding/json"
    "bytes"
)

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// Behavior matches encoding/json with UseNumber enabled.
func Unmarshal(data []byte, v any) error {
    dec := stdjson.NewDecoder(bytes.NewReader(data))
    dec.UseNumber()
    return dec.Decode(v)
}

// UnmarshalStrict is like Unmarshal but fails on unknown object fields.
func UnmarshalStrict(data []byte, v any) error {
    dec := stdjson.NewDecoder(bytes.NewReader(data))
    dec.UseNumber()
    dec.DisallowUnknownFields()
    return dec.Decode(v)
}

// Canonicalize re-encodes arbitrary JSON with deterministic key ordering.
func Canonicalize(data []byte) ([]byte, error) {
    var in any
    dec := stdjson.NewDecoder(bytes.NewReader(data))
    dec.UseNumber()
    if err := dec.Decode(&in); err != nil { return nil, err }
    var out bytes.Buffer
    if err := encodeCanonical(&out, in); err != nil { return nil, err }
    return out.Bytes(), nil
}

// Compact minifies a JSON document without changing determinism semantics.
func Compact(data []byte) ([]byte, error) {
    var out bytes.Buffer
    if err := stdjson.Compact(&out, data); err != nil { return nil, err }
    return out.Bytes(), nil
}

// EqualCanonical reports whether two JSON documents are equivalent after canonicalization.
func EqualCanonical(a, b []byte) (bool, error) {
    ca, err := Canonicalize(a)
    if err != nil { return false, err }
    cb, err := Canonicalize(b)
    if err != nil { return false, err }
    return bytes.Equal(ca, cb), nil
}

// Valid reports whether data is a syntactically valid JSON document.
func Valid(data []byte) bool { return stdjson.Valid(data) }
