package logging

import (
    logschema "github.com/sam-caldwell/ami/src/schemas/log"
)

// JSONFormatter emits deterministic log.v1 JSON lines.
type JSONFormatter struct{}

func (f JSONFormatter) Format(r Record) []byte {
    rec := logschema.Record{
        Timestamp: r.Timestamp,
        Level:     logschema.Level(r.Level),
        Message:   normalizeMsg(r.Message),
        Package:   r.Package,
        Fields:    r.Fields,
    }
    b, _ := rec.MarshalJSON()
    // newline-delimited JSON for streaming
    return append(b, '\n')
}
