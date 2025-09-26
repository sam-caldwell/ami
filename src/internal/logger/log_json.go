package logger

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

func (l *Logger) logJSON(level, msg string, data map[string]interface{}) {
    rec := record{
        Schema:    "diag.v1",
        Timestamp: FormatTimestamp(time.Now()),
        Level:     level,
        Message:   msg,
        Data:      data,
    }
    b, _ := json.Marshal(rec)
    // JSON output always goes to stdout
    fmt.Fprintln(os.Stdout, string(b))
}

