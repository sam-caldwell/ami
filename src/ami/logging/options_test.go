package logging

import "testing"

func TestOptions_ZeroValue(t *testing.T) {
    var o Options
    if o.JSON || o.Verbose || o.Color || o.Package != "" || o.Out != nil || o.DebugDir != "" {
        t.Fatalf("unexpected zero value: %+v", o)
    }
}

