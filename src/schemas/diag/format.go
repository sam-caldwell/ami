package diag

// Line encodes the record to JSON and appends a newline for NDJSON streaming.
func Line(r Record) []byte {
    b, _ := r.MarshalJSON()
    return append(b, '\n')
}

