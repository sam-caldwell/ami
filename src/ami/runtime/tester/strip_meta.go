package tester

import "strings"

func stripMeta(v any) any {
    m, ok := v.(map[string]any)
    if !ok {
        return v
    }
    out := map[string]any{}
    for k, val := range m {
        kl := strings.ToLower(k)
        if kl == "sleep_ms" || kl == "error_code" ||
            kl == "kv_pipeline" || kl == "kv_node" ||
            kl == "kv_put_key" || kl == "kv_put_val" ||
            kl == "kv_get_key" || kl == "kv_emit" {
            continue
        }
        out[k] = val
    }
    return out
}

