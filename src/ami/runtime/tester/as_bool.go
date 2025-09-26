package tester

import "strings"

func asBool(v any) (bool, bool) {
    switch t := v.(type) {
    case bool:
        return t, true
    case string:
        switch strings.ToLower(strings.TrimSpace(t)) {
        case "true", "1", "yes", "y":
            return true, true
        case "false", "0", "no", "n":
            return false, true
        }
    }
    return false, false
}

