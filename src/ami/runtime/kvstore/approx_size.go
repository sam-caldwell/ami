package kvstore

import (
    "encoding/json"
    "fmt"
)

// approxSize computes an approximate size of v in bytes by JSON encoding,
// falling back to fmt.Sprintf when necessary.
func approxSize(v any) int64 {
    if v == nil {
        return 0
    }
    switch t := v.(type) {
    case string:
        return int64(len(t))
    case []byte:
        return int64(len(t))
    default:
        if b, err := json.Marshal(t); err == nil {
            return int64(len(b))
        }
        return int64(len(fmt.Sprintf("%v", v)))
    }
}

