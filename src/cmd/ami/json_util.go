package main

import (
    "encoding/json"
    "os"
)

// writeJSONFile writes v as JSON to path with 0644 permissions.
func writeJSONFile(path string, v any) error {
    f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
    if err != nil { return err }
    enc := json.NewEncoder(f)
    enc.SetEscapeHTML(false)
    err = enc.Encode(v)
    cerr := f.Close()
    if err != nil { return err }
    return cerr
}

