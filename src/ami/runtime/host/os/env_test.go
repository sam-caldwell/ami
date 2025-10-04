package os

import (
    goos "os"
    "strings"
    "testing"
)

func TestEnv_GetSetList(t *testing.T) {
    name := "AMI_TEST_ENV_" + strings.ReplaceAll(goos.Getenv("RANDOM"), " ", "")
    if name == "AMI_TEST_ENV_" { name = "AMI_TEST_ENV_STATIC" }

    // Ensure unset or cleared baseline
    _ = goos.Unsetenv(name)
    if val := GetEnv(name); val != "" { t.Fatalf("expected empty before set, got %q", val) }

    if err := SetEnv(name, "value1"); err != nil { t.Fatalf("setenv: %v", err) }
    if val := GetEnv(name); val != "value1" { t.Fatalf("GetEnv got %q want value1", val) }

    if err := SetEnv(name, "value2"); err != nil { t.Fatalf("setenv 2: %v", err) }
    if val := GetEnv(name); val != "value2" { t.Fatalf("GetEnv got %q want value2", val) }

    // ListEnv should contain the name
    found := false
    for _, k := range ListEnv() { if k == name { found = true; break } }
    if !found { t.Fatalf("ListEnv missing %s", name) }
}
