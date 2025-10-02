package driver

// small helper for human one-liners without extra deps
func joinCSV(ss []string) string {
    if len(ss) == 0 { return "" }
    out := ss[0]
    for i := 1; i < len(ss); i++ { out += "," + ss[i] }
    return out
}

