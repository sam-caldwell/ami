package main

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }

