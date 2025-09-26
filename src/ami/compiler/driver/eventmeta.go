package driver

import sch "github.com/sam-caldwell/ami/src/schemas"

// buildEventMeta constructs a basic EventMetaV1 scaffold for debug artifacts.
func buildEventMeta(pkgName, file string) sch.EventMetaV1 {
    return sch.EventMetaV1{Schema: "eventmeta.v1", Package: pkgName, File: file, ImmutablePayload: true,
        Fields: []sch.EventMetaFieldV1{
            {Name: "id", Type: "string", Note: "unique event identifier"},
            {Name: "timestamp", Type: "iso8601", Note: "creation time in UTC"},
            {Name: "attempt", Type: "int", Note: "delivery/retry attempt count"},
        },
        Trace: &sch.TraceContextV1{
            Traceparent: sch.EventMetaFieldV1{Name: "traceparent", Type: "string", Note: "W3C traceparent header (version-traceid-spanid-flags)"},
            Tracestate:  sch.EventMetaFieldV1{Name: "tracestate", Type: "string", Note: "W3C tracestate header (vendor extensions)"},
        },
    }
}

