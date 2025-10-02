package merge

import "time"

func toTime(v any) (time.Time, bool) {
    switch x := v.(type) {
    case time.Time: return x, true
    case string:
        // try RFC3339 subset
        if t, err := time.Parse(time.RFC3339, x); err == nil { return t, true }
        // try unix seconds
        return time.Unix(0,0), false
    default:
        return time.Unix(0,0), false
    }
}

