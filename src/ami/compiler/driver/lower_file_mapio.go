package driver

import "strings"

// mapIOCapability converts an io.* step name into a coarse-grained capability string.
func mapIOCapability(name string) string {
    // name like "io.Read", "io.WriteFile", "io.Open", "io.Connect", etc.
    s := strings.TrimPrefix(name, "io.")
    if s == name || s == "" { return "io" }
    // take leading identifier portion
    end := len(s)
    for i := 0; i < len(s); i++ {
        if s[i] < 'A' || (s[i] > 'Z' && s[i] < 'a') || s[i] > 'z' { end = i; break }
    }
    head := s[:end]
    switch strings.ToLower(head) {
    case "read", "readfile", "recv":
        return "io.read"
    case "write", "writefile", "send":
        return "io.write"
    case "open", "close":
        return "io.fs"
    case "connect", "listen", "dial":
        return "network"
    default:
        return "io"
    }
}

