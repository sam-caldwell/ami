package sem

import "strings"

// ioAllowedIngress returns true if the io.* operation is permitted in an ingress position.
// Allowed families: Stdin, NetListen (listen/bind/accept), FileRead/Open(read-only),
// DirectoryList, FileStat, FileSeek.
func ioAllowedIngress(op string) bool {
    s := strings.ToLower(op)
    if strings.Contains(s, ".stdin") { return true }
    if strings.Contains(s, ".listen") || strings.Contains(s, ".bind") || strings.Contains(s, ".accept") { return true }
    if strings.Contains(s, ".read") && !strings.Contains(s, ".readwrite") { return true }
    if strings.Contains(s, ".open") && !strings.Contains(s, ".write") { return true }
    if strings.Contains(s, ".ls") || strings.Contains(s, ".listdir") || strings.Contains(s, ".readdir") || strings.Contains(s, ".dirlist") { return true }
    if strings.Contains(s, ".stat") { return true }
    if strings.Contains(s, ".seek") { return true }
    return false
}

