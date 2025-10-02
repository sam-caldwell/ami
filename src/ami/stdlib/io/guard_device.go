package io

func guardDevice() error { if !current.AllowDevice { return ErrCapabilityDenied }; return nil }

