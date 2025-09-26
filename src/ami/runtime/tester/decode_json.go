package tester

import (
    "encoding/json"
    "strings"
)

func decodeJSON(s string) (any, error) {
    dec := json.NewDecoder(strings.NewReader(s))
    dec.UseNumber()
    var v any
    if err := dec.Decode(&v); err != nil {
        return nil, err
    }
    return v, nil
}

