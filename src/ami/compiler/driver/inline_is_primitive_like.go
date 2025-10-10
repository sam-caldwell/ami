package driver

import "strings"

func isPrimitiveLike(t string) bool {
    tt := strings.TrimSpace(t)
    switch tt {
    case "bool", "int", "int64", "uint64", "real", "float64", "string":
        return true
    default:
        return false
    }
}

