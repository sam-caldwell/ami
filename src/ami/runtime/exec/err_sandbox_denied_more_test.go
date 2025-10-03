package exec

import "testing"

func TestErrSandboxDenied_ErrorString(t *testing.T) {
    e := ErrSandboxDenied{Cap: "device"}
    if e.Error() == "" { t.Fatalf("empty error string") }
}

