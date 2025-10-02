package gpu

import (
    "os"
    "testing"
)

func TestEnvBool_FilePair(t *testing.T) {
    _ = os.Setenv("AMI_GPU_TEST_BOOL", "true")
    if !envBoolTrue("AMI_GPU_TEST_BOOL") { t.Fatalf("expected true") }
    _ = os.Unsetenv("AMI_GPU_TEST_BOOL")
}

