package io

// current policy (package-scoped)
var current Policy = Policy{AllowFS: true, AllowNet: true, AllowDevice: true}

// SetPolicy sets the current I/O capability policy (global, test/runtime scoped).
func SetPolicy(p Policy) { current = p }

