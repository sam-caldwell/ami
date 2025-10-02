package main

import "strings"

// ioAllowedIngress returns true if the io.* operation is permitted in an ingress node
// per capability families: Stdin, FileRead, DirectoryList, FileStat, FileSeek.
func ioAllowedIngress(op string) bool {
    // normalize
    s := strings.ToLower(op)
    // Stdin
    if strings.Contains(s, ".stdin") { return true }
    // NetListen (ingress source): listen/bind/accept families
    if strings.Contains(s, ".listen") || strings.Contains(s, ".bind") || strings.Contains(s, ".accept") {
        return true
    }
    // FileRead
    if strings.Contains(s, ".read") && !strings.Contains(s, ".readwrite") { return true }
    if strings.Contains(s, ".open") && !strings.Contains(s, ".write") { return true }
    // DirectoryList
    if strings.Contains(s, ".ls") || strings.Contains(s, ".listdir") || strings.Contains(s, ".readdir") || strings.Contains(s, ".dirlist") { return true }
    // FileStat
    if strings.Contains(s, ".stat") { return true }
    // FileSeek
    if strings.Contains(s, ".seek") { return true }
    return false
}

