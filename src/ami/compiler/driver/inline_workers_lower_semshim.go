package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/sem"

// semInlineSig is a tiny shim for calling sem.inlineFuncSig from driver.
func semInlineSig(s string) (string, []string, bool) { return sem.InlineFuncSigForDriver(s) }

