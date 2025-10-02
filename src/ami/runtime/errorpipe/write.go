package errorpipe

import (
    "encoding/json"
    "io"
    "time"

    serr "github.com/sam-caldwell/ami/src/schemas/errors"
)

// Write writes a single errors.v1 record to w. It is safe for the default
// ErrorPipeline implementation which writes runtime errors to stderr.
func Write(w io.Writer, code, message, file string, data map[string]any) error {
    rec := serr.Error{Timestamp: time.Now().UTC(), Level: "error", Code: code, Message: message, File: file}
    if len(data) > 0 { rec.Data = data }
    b, err := json.Marshal(rec)
    if err != nil { return err }
    b = append(b, '\n')
    _, err = w.Write(b)
    return err
}

