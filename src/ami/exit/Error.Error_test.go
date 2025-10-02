package exit

import "testing"

// TestError_ErrorMethod_ReturnsMsg ensures the error string equals Msg.
func TestError_ErrorMethod_ReturnsMsg(t *testing.T) {
    // Non-empty message
    e := Error{Code: Internal, Msg: "boom"}
    if got := e.Error(); got != "boom" {
        t.Fatalf("Error()=%q, want %q", got, "boom")
    }

    // Empty message (zero-value string)
    var empty Error
    if got := empty.Error(); got != "" {
        t.Fatalf("Error() for zero-value=%q, want empty string", got)
    }
}
