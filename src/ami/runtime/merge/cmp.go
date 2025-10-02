package merge

import "time"

func cmp(a, b any) int {
    switch av := a.(type) {
    case bool:
        // false < true
        bv, ok := b.(bool); if !ok { return 0 }
        if !av && bv { return -1 }
        if av && !bv { return 1 }
        return 0
    case int:
        bv, ok := b.(int); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case int64:
        switch bv := b.(type) {
        case int64:
            if av < bv {return -1} else if av > bv {return 1}; return 0
        case int:
            ai := av; bi := int64(bv)
            if ai < bi {return -1} else if ai > bi {return 1}; return 0
        default:
            return 0
        }
    case float64:
        bv, ok := b.(float64); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case string:
        bv, ok := b.(string); if !ok { return 0 }
        switch { case av < bv: return -1; case av > bv: return 1; default: return 0 }
    case time.Time:
        bv, ok := b.(time.Time); if !ok { return 0 }
        switch { case av.Before(bv): return -1; case av.After(bv): return 1; default: return 0 }
    default:
        return 0
    }
}

