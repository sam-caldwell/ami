package sem

import (
    "time"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeMultiPath validates MultiPath usage and merge.* attributes on Collect steps.
// - edge.MultiPath/MultiPath only valid on Collect → E_MP_ONLY_COLLECT
// - merge.* unknown → E_MERGE_ATTR_UNKNOWN
// - merge.* invalid arity → E_MERGE_ATTR_ARGS
// - merge.Sort without a field → W_MERGE_SORT_NO_FIELD
func AnalyzeMultiPath(f *ast.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    now := time.Unix(0, 0).UTC()
    // allowed merge attributes and arity constraints (min,max; -1 means unbounded)
    type ar struct{ min, max int }
    merges := map[string]ar{
        "merge.Sort":       {1, 2},
        "merge.Stable":     {0, 0},
        "merge.Key":        {1, 1},
        "merge.Dedup":      {0, 1},
        "merge.Window":     {1, 1},
        "merge.Watermark":  {2, 2},
        "merge.Timeout":    {1, 1},
        "merge.Buffer":     {1, 2},
        "merge.PartitionBy":{1, 1},
    }
    for _, d := range f.Decls {
        pd, ok := d.(*ast.PipelineDecl)
        if !ok { continue }
        for _, s := range pd.Stmts {
            st, ok := s.(*ast.StepStmt)
            if !ok { continue }
            // detect multipath on non-Collect
            for _, at := range st.Attrs {
                if at.Name == "edge.MultiPath" || at.Name == "MultiPath" {
                    if st.Name != "Collect" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MP_ONLY_COLLECT", Message: "edge.MultiPath only valid on Collect nodes", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"attr": at.Name, "step": st.Name}})
                    }
                }
            }
            if st.Name != "Collect" { continue }
            // validate merge.* attributes and basic normalization
            seen := map[string]string{}
            keyField := ""
            partitionField := ""
            hasSort := false
            hasStable := false
            hasWindow := false
            hasWatermark := false
            dedupNoField := false
            var sortFields []string
            watermarkField := ""
            for _, at := range st.Attrs {
                if strings.HasPrefix(at.Name, "merge.") {
                    if rng, ok := merges[at.Name]; ok {
                        argc := len(at.Args)
                        if at.Name == "merge.Sort" && argc == 0 {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"attr": at.Name}})
                            continue
                        }
                        if argc < rng.min || (rng.max >= 0 && argc > rng.max) {
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: at.Name + ": invalid number of arguments", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"argc": argc, "expected_min": rng.min, "expected_max": rng.max}})
                            continue
                        }
                        // additional validations
                        switch at.Name {
                        case "merge.Sort":
                            hasSort = true
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: "merge.Sort: field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"attr": at.Name}})
                            }
                            // basic field name validation: [A-Za-z_][A-Za-z0-9_\.]*
                            if argc >= 1 {
                                fld := at.Args[0].Text
                                if !validFieldName(fld) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_FIELD_NAME_INVALID", Message: "merge.Sort: invalid field name", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"field": fld}})
                                }
                                sortFields = append(sortFields, fld)
                            }
                            if argc >= 2 {
                                ord := at.Args[1].Text
                                if ord != "asc" && ord != "desc" {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_ORDER_INVALID", Message: "merge.Sort: order must be 'asc' or 'desc'", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"order": ord}})
                                }
                            }
                        case "merge.Stable":
                            hasStable = true
                        case "merge.Watermark":
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: "merge.Watermark: field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"attr": at.Name}})
                            }
                            if argc >= 2 {
                                lat := strings.TrimSpace(at.Args[1].Text)
                                // Determine format: integer millis or duration with unit.
                                // Classify: malformed → E_MERGE_ATTR_TYPE; non‑positive but well‑formed → warn.
                                if isInteger(lat) {
                                    if !validNonNegativeInt(lat) {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Watermark: lateness must be positive int or duration (e.g., 100ms,1s,2m,1h)", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"lateness": lat}})
                                    } else if lat == "0" {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WATERMARK_NONPOSITIVE", Message: "merge.Watermark: lateness should be > 0", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"lateness": lat}})
                                    }
                                } else if isDurationLike(lat) {
                                    // extract numeric prefix then validate non‑negative
                                    num := numericPrefix(lat)
                                    if num == "" {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Watermark: lateness must be positive int or duration (e.g., 100ms,1s,2m,1h)", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"lateness": lat}})
                                    } else if num == "0" {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WATERMARK_NONPOSITIVE", Message: "merge.Watermark: lateness should be > 0", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"lateness": lat}})
                                    }
                                } else {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Watermark: lateness must be positive int or duration (e.g., 100ms,1s,2m,1h)", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"lateness": lat}})
                                }
                            }
                            // validate field name when present
                            if argc >= 1 {
                                fld := at.Args[0].Text
                                if !validFieldName(fld) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_FIELD_NAME_INVALID", Message: "merge.Watermark: invalid field name", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"field": fld}})
                                }
                                watermarkField = fld
                            }
                            hasWatermark = true
                        case "merge.Window":
                            if argc >= 1 {
                                if !validNonNegativeInt(at.Args[0].Text) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Window: size must be a non-negative integer", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"size": strings.TrimSpace(at.Args[0].Text)}})
                                } else if at.Args[0].Text == "0" || strings.HasPrefix(at.Args[0].Text, "-") {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WINDOW_ZERO_OR_NEGATIVE", Message: "merge.Window: size should be > 0", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"size": strings.TrimSpace(at.Args[0].Text)}})
                                }
                            }
                            hasWindow = true
                        case "merge.Timeout":
                            if argc >= 1 {
                                ms := strings.TrimSpace(at.Args[0].Text)
                                // First, ensure it is an integer at all → type error if not.
                                if !isInteger(ms) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Timeout: must be a positive integer (ms)", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"ms": ms}})
                                } else {
                                    // Parsed as integer; classify non‑positive as ARGS error per spec/tests.
                                    if !validPositiveInt(ms) {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Timeout: must be > 0", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"ms": ms}})
                                    }
                                }
                            }
                        case "merge.Key", "merge.PartitionBy":
                            if argc >= 1 && strings.TrimSpace(at.Args[0].Text) == "" {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_REQUIRED", Message: at.Name + ": field is required", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"attr": at.Name}})
                            }
                            if argc >= 1 {
                                fld := at.Args[0].Text
                                if !validFieldName(fld) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_FIELD_NAME_INVALID", Message: at.Name + ": invalid field name", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"field": fld}})
                                }
                            }
                        case "merge.Buffer":
                            if argc >= 1 {
                                cap := strings.TrimSpace(at.Args[0].Text)
                                if !validNonNegativeInt(cap) {
                                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_TYPE", Message: "merge.Buffer: capacity must be a non-negative integer", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"capacity": cap}})
                                }
                                if cap == "0" || cap == "1" {
                                    if argc >= 2 {
                                        pol := at.Args[1].Text
                                        if pol == "dropOldest" || pol == "dropNewest" {
                                            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_TINY_BUFFER", Message: "merge.Buffer: tiny capacity with dropping policy", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"capacity": cap, "policy": pol}})
                                        }
                                    }
                                }
                                if argc >= 2 {
                                    pol := at.Args[1].Text
                                    if pol == "drop" {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_BUFFER_DROP_ALIAS", Message: "merge.Buffer: ambiguous 'drop' alias; use dropOldest/dropNewest/block", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"policy": pol}})
                                    } else if pol != "block" && pol != "dropOldest" && pol != "dropNewest" {
                                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_ARGS", Message: "merge.Buffer: policy must be one of block|dropOldest|dropNewest", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"policy": pol}})
                                    }
                                }
                            }
                        }
                        // track fields for combo checks
                        if at.Name == "merge.Key" && argc >= 1 { keyField = at.Args[0].Text }
                        if at.Name == "merge.PartitionBy" && argc >= 1 { partitionField = at.Args[0].Text }
                        if at.Name == "merge.Dedup" && argc == 0 { dedupNoField = true }
                        // conflict detection on repeated attributes with differing normalized value
                        key := at.Name
                        val := canonicalAttrValue(at.Name, at.Args)
                        if prev, ok := seen[key]; ok {
                            if prev != val {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: at.Name + ": conflicting values", Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"prev": prev, "value": val}})
                            }
                        } else {
                            seen[key] = val
                        }
                    } else {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_UNKNOWN", Message: "unknown merge attribute: " + at.Name, Pos: &diag.Position{Line: at.Pos.Line, Column: at.Pos.Column, Offset: at.Pos.Offset}, Data: map[string]any{"name": at.Name}})
                    }
                }
            }
            // cross-attribute conflicts
            if keyField != "" && partitionField != "" && keyField != partitionField {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: "merge.PartitionBy vs merge.Key conflict", Pos: &p, Data: map[string]any{"key": keyField, "partition": partitionField}})
            }
            // Require primary Sort field to match Key when both provided.
            if keyField != "" && hasSort {
                // primary is the first merge.Sort encountered
                primary := ""
                if len(sortFields) > 0 { primary = sortFields[0] }
                if primary != keyField {
                    p := stepPos(st)
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_PRIMARY_NEQ_KEY", Message: "merge.Sort primary must equal merge.Key", Pos: &p, Data: map[string]any{"key": keyField, "primary": primary, "sort": sortFields}})
                }
            }
            // If PartitionBy is present without Key, require primary Sort to match PartitionBy.
            if keyField == "" && partitionField != "" && hasSort {
                primary := ""
                if len(sortFields) > 0 { primary = sortFields[0] }
                if primary != partitionField {
                    p := stepPos(st)
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_PRIMARY_NEQ_PARTITION", Message: "merge.Sort primary must equal merge.PartitionBy when Key is not set", Pos: &p, Data: map[string]any{"partition": partitionField, "primary": primary, "sort": sortFields}})
                }
            }
            // stable requested but no sort field specified
            if hasStable && !hasSort {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_STABLE_WITHOUT_SORT", Message: "merge.Stable has no effect without merge.Sort", Pos: &p, Data: map[string]any{"stable": true}})
            }
            // window without watermark hint
            if hasWindow && !hasWatermark {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WINDOW_WITHOUT_WATERMARK", Message: "merge.Window without Watermark may be based on processing time", Pos: &p, Data: map[string]any{"window": true}})
            }
            // watermark field not used as primary sort
            if hasWatermark && hasSort {
                primary := ""
                if len(sortFields) > 0 { primary = sortFields[0] }
                if primary != "" && watermarkField != "" && primary != watermarkField {
                    p := stepPos(st)
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_WATERMARK_NOT_PRIMARY_SORT", Message: "merge.Watermark field is not the primary merge.Sort field", Pos: &p, Data: map[string]any{"watermark": watermarkField, "primary": primary, "sort": sortFields}})
                }
            }
            // sort stability hints: sort without key/partition and without stable may be unstable across batches
            if hasSort && !hasStable && keyField == "" && partitionField == "" {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_POSSIBLY_UNSTABLE", Message: "merge.Sort without Key/Partition and Stable may be unstable", Pos: &p, Data: map[string]any{"fields": sortFields}})
            }
            // redundant stable: when a unique key is present, Stable often provides no additional guarantees
            if hasStable && keyField != "" && hasSort {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Info, Code: "W_MERGE_STABLE_REDUNDANT", Message: "merge.Stable may be redundant when a unique Key is present", Pos: &p, Data: map[string]any{"key": keyField, "sort": sortFields}})
            }
            // Dedup(field) conflicts with Key when both provided and different; also hint under PartitionBy without Key.
            for _, at := range st.Attrs {
                if at.Name == "merge.Dedup" && len(at.Args) >= 1 {
                    df := strings.TrimSpace(at.Args[0].Text)
                    if df != "" {
                        if keyField != "" && df != keyField {
                            p := stepPos(st)
                            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_ATTR_CONFLICT", Message: "merge.Dedup field differs from merge.Key", Pos: &p, Data: map[string]any{"dedup": df, "key": keyField}})
                        }
                        if keyField == "" && partitionField != "" {
                            p := stepPos(st)
                            if StrictDedupUnderPartition {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION", Message: "merge.Dedup(field) under PartitionBy without Key may mis-deduplicate across partitions", Pos: &p, Data: map[string]any{"dedup": df, "partitionBy": partitionField}})
                            } else {
                                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION", Message: "merge.Dedup(field) under PartitionBy without Key may be ineffective", Pos: &p, Data: map[string]any{"dedup": df, "partitionBy": partitionField}})
                            }
                        }
                    }
                }
            }
            // Dedup() without explicit field relies on merge.Key; warn if neither provided.
            if dedupNoField && keyField == "" {
                p := stepPos(st)
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_DEDUP_WITHOUT_KEY", Message: "merge.Dedup without field requires merge.Key", Pos: &p, Data: map[string]any{"dedup": true}})
            }
            // Dedup() without key under partitioning may not deduplicate as expected across partitions
            if dedupNoField && partitionField != "" && keyField == "" {
                p := stepPos(st)
                if StrictDedupUnderPartition {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_DEDUP_WITHOUT_KEY_UNDER_PARTITION", Message: "merge.Dedup without key under PartitionBy may be ineffective", Pos: &p, Data: map[string]any{"partitionBy": partitionField}})
                } else {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_DEDUP_WITHOUT_KEY_UNDER_PARTITION", Message: "merge.Dedup without key under PartitionBy may be ineffective", Pos: &p, Data: map[string]any{"partitionBy": partitionField}})
                }
            }
        }
    }
    return out
}
