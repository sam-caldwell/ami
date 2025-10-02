package amsignal

func safeCall(f func()) { defer func(){ _ = recover() }(); if f != nil { f() } }

