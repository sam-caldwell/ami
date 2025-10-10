package sem

// InlineFuncSigForDriver exposes inlineFuncSig for driver usage without duplicating logic.
func InlineFuncSigForDriver(text string) (string, []string, bool) { return inlineFuncSig(text) }

