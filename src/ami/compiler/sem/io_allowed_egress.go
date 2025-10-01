package sem

import "strings"

// ioAllowedEgress returns true if the io.* operation is permitted in an egress position.
// Allowed families: Stdout, Stderr, NetConnect (connect/dial), Net*Send(send/sendto),
// FileWrite/Append/Create/Delete/Truncate/Chmod/Chown, FileStat/Read/Seek,
// DirectoryCreate/Delete, Temp file/dir creation.
func ioAllowedEgress(op string) bool {
    s := strings.ToLower(op)
    if strings.Contains(s, ".stdout") || strings.Contains(s, ".stderr") { return true }
    if strings.Contains(s, ".connect") || strings.Contains(s, ".dial") { return true }
    if strings.Contains(s, ".send") || strings.Contains(s, ".sendto") {
        if strings.Contains(s, "udp") || strings.Contains(s, "tcp") || strings.Contains(s, "icmp") { return true }
        return true
    }
    if strings.Contains(s, ".write") || strings.Contains(s, ".append") { return true }
    if strings.Contains(s, ".create") || strings.Contains(s, ".delete") || strings.Contains(s, ".truncate") || strings.Contains(s, ".chmod") || strings.Contains(s, ".chown") { return true }
    if strings.Contains(s, ".stat") || strings.Contains(s, ".read") || strings.Contains(s, ".seek") { return true }
    if strings.Contains(s, ".mkdir") || strings.Contains(s, ".mkdirall") || strings.Contains(s, ".dircreate") || strings.Contains(s, ".rmdir") || strings.Contains(s, ".dirdelete") { return true }
    if strings.Contains(s, ".tempfile") || strings.Contains(s, ".createtemp") || strings.Contains(s, ".tempdir") || strings.Contains(s, ".createtempdir") { return true }
    return false
}

