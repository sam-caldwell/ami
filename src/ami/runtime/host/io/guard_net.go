package io

func guardNet() error { if !current.AllowNet { return ErrCapabilityDenied }; return nil }

