package driver

import "strings"

func isNumericLike(t string) bool {
    tt := strings.TrimSpace(t)
    switch tt {
    case "int", "int64", "uint64", "real", "float64":
        return true
    default:
        return false
    }
}

