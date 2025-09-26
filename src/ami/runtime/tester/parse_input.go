package tester

import (
    "strings"
)

func parseInput(s string) (any, metaInfo, error) {
    if strings.TrimSpace(s) == "" {
        return nil, metaInfo{}, nil
    }
    v, err := decodeJSON(s)
    if err != nil {
        return nil, metaInfo{}, err
    }
    // Extract meta keys when object
    mi := metaInfo{}
    if m, ok := v.(map[string]any); ok {
        if n, ok := asInt(m["sleep_ms"]); ok {
            mi.sleepMs = n
        }
        if ec, ok := m["error_code"].(string); ok {
            mi.errorCode = ec
        }
        if p, ok := m["kv_pipeline"].(string); ok {
            mi.kvPipeline = p
        }
        if n, ok := m["kv_node"].(string); ok {
            mi.kvNode = n
        }
        if k, ok := m["kv_put_key"].(string); ok {
            mi.kvPutKey = k
        }
        if pv, ok := m["kv_put_val"]; ok {
            mi.kvPutVal = pv
        }
        if gk, ok := m["kv_get_key"].(string); ok {
            mi.kvGetKey = gk
        }
        if b, ok := asBool(m["kv_emit"]); ok {
            mi.kvEmit = b
        }
    }
    return v, mi, nil
}

