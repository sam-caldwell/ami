package main

import "strings"

// ioAllowedEgress returns true if the io.* operation is permitted in an egress node
// per capability families: Stdout, Stderr, FileWrite, FileCreate, FileDelete, FileTruncate,
// FileAppend, FileChmod, FileStat, DirectoryCreate, DirectoryDelete, TempFileCreate,
// TempDirectoryCreate, FileRead, FileChown, FileSeek.
func ioAllowedEgress(op string) bool {
    s := strings.ToLower(op)
    // Stdout/Stderr
    if strings.Contains(s, ".stdout") || strings.Contains(s, ".stderr") { return true }
    // NetConnect (egress sink): connect/dial families
    if strings.Contains(s, ".connect") || strings.Contains(s, ".dial") { return true }
    // NetUdpSend / NetTcpSend / NetIcmpSend (egress sink): send families with protocol hints
    if strings.Contains(s, ".send") || strings.Contains(s, ".sendto") {
        if strings.Contains(s, "udp") || strings.Contains(s, "tcp") || strings.Contains(s, "icmp") { return true }
        // If protocol unspecified, still consider send as an egress network op
        return true
    }
    // FileWrite/Append
    if strings.Contains(s, ".write") || strings.Contains(s, ".append") { return true }
    // FileCreate/Delete/Truncate/Chmod/Chown
    if strings.Contains(s, ".create") || strings.Contains(s, ".delete") || strings.Contains(s, ".truncate") || strings.Contains(s, ".chmod") || strings.Contains(s, ".chown") { return true }
    // FileStat/Read/Seek
    if strings.Contains(s, ".stat") || strings.Contains(s, ".read") || strings.Contains(s, ".seek") { return true }
    // DirectoryCreate/Delete
    if strings.Contains(s, ".mkdir") || strings.Contains(s, ".mkdirall") || strings.Contains(s, ".dircreate") || strings.Contains(s, ".rmdir") || strings.Contains(s, ".dirdelete") { return true }
    // Temp file/dir creation
    if strings.Contains(s, ".tempfile") || strings.Contains(s, ".createtemp") || strings.Contains(s, ".tempdir") || strings.Contains(s, ".createtempdir") { return true }
    return false
}

