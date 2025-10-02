package sem

import "strings"

func keyVal(s string) (string, string) {
    in := innerGeneric(s)
    // split on first comma
    c := strings.IndexByte(in, ',')
    if c < 0 { return in, "" }
    return strings.TrimSpace(in[:c]), strings.TrimSpace(in[c+1:])
}

