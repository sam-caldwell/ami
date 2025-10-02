package amitime

func safe(f func()) { defer func(){ _ = recover() }(); if f != nil { f() } }

