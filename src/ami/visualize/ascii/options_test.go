package ascii

import "testing"

func TestOptions_StructFields_Roundtrip(t *testing.T) {
    // Zero-value check
    var o Options
    if o.Width != 0 || o.Focus != "" || o.Legend || o.Color { t.Fatalf("unexpected zero value: %+v", o) }
    // Set and verify
    o = Options{Width: 80, Focus: "Node", Legend: true, Color: true}
    if o.Width != 80 || o.Focus != "Node" || !o.Legend || !o.Color {
        t.Fatalf("unexpected values: %+v", o)
    }
}

