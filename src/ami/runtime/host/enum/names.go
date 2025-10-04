package enum

// Names returns all canonical enum names in canonical order.
func Names(d Descriptor) []string { return append([]string(nil), d.Names...) }

