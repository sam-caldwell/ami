package main

func strOrEmpty(v any) string { if s, ok := v.(string); ok { return s }; return "" }

