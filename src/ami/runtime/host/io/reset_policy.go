package io

// ResetPolicy restores the default permissive policy.
func ResetPolicy() { current = Policy{AllowFS: true, AllowNet: true, AllowDevice: true} }

