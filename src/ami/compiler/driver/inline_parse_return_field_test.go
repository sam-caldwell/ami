package driver

import "testing"

func Test_parseInlineReturn_Field(t *testing.T) {
    rp, ok := parseInlineReturn("return event.payload.field.user.id")
    if !ok || rp.kind != retField || rp.path != "user.id" {
        t.Fatalf("field parse failed: ok=%v rp=%+v", ok, rp)
    }
}

