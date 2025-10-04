package trigger

import (
    "net"
    "strconv"
)

// splitHostPort attempts to split a host:port string into parts.
func splitHostPort(addr string) (string, uint16) {
    if addr == "" { return "", 0 }
    host, portStr, err := net.SplitHostPort(addr)
    if err != nil { return addr, 0 }
    p, _ := strconv.Atoi(portStr)
    if p < 0 { p = 0 }
    if p > 65535 { p = 65535 }
    return host, uint16(p)
}

