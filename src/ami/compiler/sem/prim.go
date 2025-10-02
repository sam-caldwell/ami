package sem

func prim(t string) bool {
    switch t {
    case "", "any":
        return false
    case "bool", "byte", "rune",
        "int", "int8", "int16", "int32", "int64", "int128",
        "uint", "uint8", "uint16", "uint32", "uint64", "uint128",
        "float32", "float64",
        "string":
        return true
    default:
        return false
    }
}

