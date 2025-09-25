package json

import (
    "bytes"
    stdjson "encoding/json"
    "sort"
)

// Marshal encodes v into JSON with deterministic map key ordering.
func Marshal(v any) ([]byte, error) {
    // First marshal using the standard library to honor struct tags and custom marshalers.
    raw, err := stdjson.Marshal(v)
    if err != nil { return nil, err }

    // Decode into generic representation, preserving numbers.
    var generic any
    dec := stdjson.NewDecoder(bytes.NewReader(raw))
    dec.UseNumber()
    if err := dec.Decode(&generic); err != nil { return nil, err }

    // Re-encode with deterministic ordering.
    var buf bytes.Buffer
    if err := encodeCanonical(&buf, generic); err != nil { return nil, err }
    return buf.Bytes(), nil
}

// MarshalIndent encodes v deterministically (sorted map keys) and applies indentation.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
    b, err := Marshal(v)
    if err != nil { return nil, err }
    var out bytes.Buffer
    if err := stdjson.Indent(&out, b, prefix, indent); err != nil { return nil, err }
    return out.Bytes(), nil
}

// encodeCanonical writes a canonical JSON encoding of v to buf with sorted map keys.
func encodeCanonical(buf *bytes.Buffer, v any) error {
    switch x := v.(type) {
    case nil:
        buf.WriteString("null")
        return nil
    case bool, stdjson.Number, string:
        // Delegate primitives to std json for correct escaping/formatting.
        b, _ := stdjson.Marshal(x)
        buf.Write(b)
        return nil
    case []any:
        buf.WriteByte('[')
        for i, el := range x {
            if i > 0 { buf.WriteByte(',') }
            if err := encodeCanonical(buf, el); err != nil { return err }
        }
        buf.WriteByte(']')
        return nil
    case map[string]any:
        // Sort keys lexicographically for determinism.
        keys := make([]string, 0, len(x))
        for k := range x { keys = append(keys, k) }
        sort.Strings(keys)
        buf.WriteByte('{')
        for i, k := range keys {
            if i > 0 { buf.WriteByte(',') }
            kb, _ := stdjson.Marshal(k)
            buf.Write(kb)
            buf.WriteByte(':')
            if err := encodeCanonical(buf, x[k]); err != nil { return err }
        }
        buf.WriteByte('}')
        return nil
    default:
        // Fallback: marshal via stdlib and re-run through canonical path.
        b, err := stdjson.Marshal(x)
        if err != nil { return err }
        var y any
        dec := stdjson.NewDecoder(bytes.NewReader(b))
        dec.UseNumber()
        if err := dec.Decode(&y); err != nil { return err }
        return encodeCanonical(buf, y)
    }
}
