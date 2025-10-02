package main

func updateSumEntry(m map[string]any, name, version, sha, source string) bool {
    p, ok := m["packages"]
    if !ok { return false }
    switch t := p.(type) {
    case []any:
        for _, el := range t {
            if mm, ok := el.(map[string]any); ok {
                if strOrEmpty(mm["name"]) == name && strOrEmpty(mm["version"]) == version {
                    mm["sha256"] = sha
                    if source != "" { mm["source"] = source }
                    return true
                }
            }
        }
    case map[string]any:
        if mm, ok := t[name].(map[string]any); ok {
            mm["version"] = version
            mm["sha256"] = sha
            if source != "" { mm["source"] = source }
            t[name] = mm
            m["packages"] = t
            return true
        }
    }
    return false
}

