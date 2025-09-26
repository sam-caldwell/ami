package tester

import (
    "errors"
    "strings"
)

func validateFixtures(fx []Fixture) error {
    for _, f := range fx {
        m := strings.ToLower(strings.TrimSpace(f.Mode))
        if m != "" && m != "ro" && m != "rw" {
            return errors.New("invalid fixture mode")
        }
        if strings.TrimSpace(f.Path) == "" {
            return errors.New("invalid fixture path")
        }
    }
    return nil
}

