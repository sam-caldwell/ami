package driver

import "testing"

func TestOptions_ZeroAndSetValues(t *testing.T) {
    var o Options
    if o.Debug { t.Fatalf("zero value Debug should be false") }
    o.Debug = true
    if !o.Debug { t.Fatalf("expected Debug=true") }
}

